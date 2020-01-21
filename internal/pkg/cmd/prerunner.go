package cmd

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/jonboulle/clockwork"
	"github.com/spf13/cobra"
	"gopkg.in/square/go-jose.v2/jwt"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/update"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// PreRun is a helper class for automatically setting up Cobra PersistentPreRun commands
type PreRunner interface {
	Anonymous(command *CLICommand) func(cmd *cobra.Command, args []string) error
	Authenticated(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error
	AuthenticatedWithMDS(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error
	HasAPIKey(command *HasAPIKeyCLICommand) func(cmd *cobra.Command, args []string) error
}

// PreRun is the standard PreRunner implementation
type PreRun struct {
	UpdateClient update.Client
	CLIName      string
	Logger       *log.Logger
	Clock        clockwork.Clock
	FlagResolver FlagResolver
	Version      *version.Version
}

type CLICommand struct {
	*cobra.Command
	Client    *ccloud.Client
	MDSClient *mds.APIClient
	Config    *DynamicConfig
	Version   *version.Version
	prerunner PreRunner
}

type AuthenticatedCLICommand struct {
	*CLICommand
	Context *DynamicContext
	State   *config.ContextState
}

type HasAPIKeyCLICommand struct {
	*CLICommand
	Context *DynamicContext
}

func (a *AuthenticatedCLICommand) AuthToken() string {
	return a.State.AuthToken
}
func (a *AuthenticatedCLICommand) EnvironmentId() string {
	return a.State.Auth.Account.Id
}

func NewAuthenticatedCLICommand(command *cobra.Command, cfg *config.Config, prerunner PreRunner) *AuthenticatedCLICommand {
	cmd := &AuthenticatedCLICommand{
		CLICommand: NewCLICommand(command, cfg, prerunner),
		Context:    nil,
		State:      nil,
	}
	command.PersistentPreRunE = prerunner.Authenticated(cmd)
	cmd.Command = command
	return cmd
}

func NewAuthenticatedWithMDSCLICommand(command *cobra.Command, cfg *config.Config, prerunner PreRunner) *AuthenticatedCLICommand {
	cmd := &AuthenticatedCLICommand{
		CLICommand: NewCLICommand(command, cfg, prerunner),
		Context:    nil,
		State:      nil,
	}
	command.PersistentPreRunE = prerunner.AuthenticatedWithMDS(cmd)
	cmd.Command = command
	return cmd
}

func NewHasAPIKeyCLICommand(command *cobra.Command, cfg *config.Config, prerunner PreRunner) *HasAPIKeyCLICommand {
	cmd := &HasAPIKeyCLICommand{
		CLICommand: NewCLICommand(command, cfg, prerunner),
		Context:    nil,
	}
	command.PersistentPreRunE = prerunner.HasAPIKey(cmd)
	cmd.Command = command
	return cmd
}

func NewAnonymousCLICommand(command *cobra.Command, cfg *config.Config, prerunner PreRunner) *CLICommand {
	cmd := NewCLICommand(command, cfg, prerunner)
	command.PersistentPreRunE = prerunner.Anonymous(cmd)
	cmd.Command = command
	return cmd
}

func NewCLICommand(command *cobra.Command, cfg *config.Config, prerunner PreRunner) *CLICommand {
	return &CLICommand{
		Config:    NewDynamicConfig(cfg, nil, nil),
		Command:   command,
		prerunner: prerunner,
	}
}

func (a *AuthenticatedCLICommand) AddCommand(command *cobra.Command) {
	command.PersistentPreRunE = a.prerunner.Authenticated(a)
	a.Command.AddCommand(command)
}

func (h *HasAPIKeyCLICommand) AddCommand(command *cobra.Command) {
	command.PersistentPreRunE = h.prerunner.HasAPIKey(h)
	h.Command.AddCommand(command)
}

// Anonymous provides PreRun operations for commands that may be run without a logged-in user
func (r *PreRun) Anonymous(command *CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		command.Version = r.Version
		command.Config.Resolver = r.FlagResolver
		if err := log.SetLoggingVerbosity(cmd, r.Logger); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if err := r.notifyIfUpdateAvailable(cmd, r.CLIName, command.Version.Version); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if command != nil {
			err := setClients(command, command.Version)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
		}
		return nil
	}
}

// Authenticated provides PreRun operations for commands that require a logged-in Confluent Cloud user.
func (r *PreRun) Authenticated(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := r.Anonymous(command.CLICommand)(cmd, args)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		ctx, err := command.Config.Context(cmd)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if ctx == nil {
			return errors.HandleCommon(errors.ErrNoContext, cmd)
		}
		command.Context = ctx
		command.State, err = ctx.AuthenticatedState(cmd)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		// Validate token (not expired)
		authToken := command.AuthToken()
		return r.validateToken(cmd, authToken)
	}
}

// Authenticated provides PreRun operations for commands that require a logged-in Confluent Cloud user.
func (r *PreRun) AuthenticatedWithMDS(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := r.Anonymous(command.CLICommand)(cmd, args)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		ctx, err := command.Config.Context(cmd)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if ctx == nil {
			return errors.HandleCommon(errors.ErrNoContext, cmd)
		}
		command.Context = ctx
		if !ctx.HasMDSLogin() {
			return errors.HandleCommon(errors.ErrNotLoggedIn, cmd)
		}
		command.State = ctx.State
		// Validate token (not expired)
		authToken := command.AuthToken()
		return r.validateToken(cmd, authToken)
	}
}

// HasAPIKey provides PreRun operations for commands that require an API key.
func (r *PreRun) HasAPIKey(command *HasAPIKeyCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := r.Anonymous(command.CLICommand)(cmd, args)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		ctx, err := command.Config.Context(cmd)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		command.Context = ctx
		hasAPIKey, err := ctx.HasAPIKey(cmd, ctx.Kafka)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if !hasAPIKey {
			err = &errors.UnspecifiedAPIKeyError{ClusterID: ctx.Kafka}
			return errors.HandleCommon(err, cmd)
		}
		return nil
	}
}

// notifyIfUpdateAvailable prints a message if an update is available
func (r *PreRun) notifyIfUpdateAvailable(cmd *cobra.Command, name string, currentVersion string) error {
	updateAvailable, _, err := r.UpdateClient.CheckForUpdates(name, currentVersion, false)
	if err != nil {
		// This is a convenience helper to check-for-updates before arbitrary commands. Since the CLI supports running
		// in internet-less environments (e.g., local or on-prem deploys), swallow the error and log a warning.
		r.Logger.Warn(err)
		return nil
	}
	if updateAvailable {
		msg := "Updates are available for %s. To install them, please run:\n$ %s update\n\n"
		ErrPrintf(cmd, msg, name, name)
	}
	return nil
}

func setClients(cliCmd *CLICommand, ver *version.Version) error {
	ctx, err := cliCmd.Config.Context(cliCmd.Command)
	if err != nil {
		return err
	}
	ccloudClient, err := createCCloudClient(ctx, cliCmd.Command, ver)
	if err != nil {
		return err
	}
	cliCmd.Client = ccloudClient
	cliCmd.MDSClient = createMDSClient(ctx, ver)
	cliCmd.Config.Client = ccloudClient
	return nil
}

func createCCloudClient(ctx *DynamicContext, cmd *cobra.Command, ver *version.Version) (*ccloud.Client, error) {
	var baseURL string
	var authToken string
	var logger *log.Logger
	var userAgent string
	if ctx != nil {
		baseURL = ctx.Platform.Server
		state, err := ctx.AuthenticatedState(cmd)
		if err != nil && err != errors.ErrNotLoggedIn {
			return nil, err
		}
		if err == nil {
			authToken = state.AuthToken
		}
		logger = ctx.Logger
		userAgent = ver.UserAgent
	}
	return ccloud.NewClientWithJWT(context.Background(), authToken, &ccloud.Params{
		BaseURL: baseURL, Logger: logger, UserAgent: userAgent,
	}), nil
}

func createMDSClient(ctx *DynamicContext, ver *version.Version) *mds.APIClient {
	mdsConfig := mds.NewConfiguration()
	if ctx != nil {
		mdsConfig.BasePath = ctx.Platform.Server
		mdsConfig.UserAgent = ver.UserAgent
	}
	return mds.NewAPIClient(mdsConfig)
}

func (r *PreRun) validateToken(cmd *cobra.Command, authToken string) error {
	// Validate token (not expired)
	var claims map[string]interface{}
	token, err := jwt.ParseSigned(authToken)
	if err != nil {
		return errors.HandleCommon(new(ccloud.InvalidTokenError), cmd)
	}
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if exp, ok := claims["exp"].(float64); ok {
		if float64(r.Clock.Now().Unix()) > exp {
			return errors.HandleCommon(new(ccloud.ExpiredTokenError), cmd)
		}
	}
	return nil
}
