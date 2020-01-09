package cmd

import (
	"context"
	"fmt"

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
	Anonymous(cfg *config.Config, command *CLICommand) func(cmd *cobra.Command, args []string) error
	Authenticated(cfg *config.Config, command *CLICommand) func(cmd *cobra.Command, args []string) error
	HasAPIKey(cfg *config.Config, command *CLICommand) func(cmd *cobra.Command, args []string) error
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
	Config    *config.Config
	Version   *version.Version
}

func NewAuthenticatedCLICommand(command *cobra.Command, cfg *config.Config, prerunner PreRunner) *CLICommand {
	cmd := &CLICommand{Config: cfg}
	command.PersistentPreRunE = prerunner.Authenticated(cfg, cmd)
	cmd.Command = command
	return cmd
}

func NewHasAPIKeyCLICommand(command *cobra.Command, cfg *config.Config, prerunner PreRunner) *CLICommand {
	cmd := &CLICommand{Config: cfg}
	command.PersistentPreRunE = prerunner.HasAPIKey(cfg, cmd)
	cmd.Command = command
	return cmd
}

func NewAnonymousCLICommand(command *cobra.Command, cfg *config.Config, prerunner PreRunner) *CLICommand {
	cmd := &CLICommand{Config: cfg}
	command.PersistentPreRunE = prerunner.Anonymous(cfg, cmd)
	cmd.Command = command
	return cmd
}

func NewCLICommand(command *cobra.Command, cfg *config.Config) *CLICommand {
	return &CLICommand{
		Config:  cfg,
		Command: command,
	}
}

// Anonymous provides PreRun operations for commands that may be run without a logged-in user
func (r *PreRun) Anonymous(cfg *config.Config, command *CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		command.Version = r.Version
		if err := log.SetLoggingVerbosity(cmd, r.Logger); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if err := r.notifyIfUpdateAvailable(cmd, r.CLIName, command.Version.Version); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if err := r.FlagResolver.ResolveContextFlag(cmd, cfg); err != nil {
			return errors.HandleCommon(err, cmd)
		}
		if command != nil {
			ctx := cfg.Context()
			setClients(command, ctx, command.Version)
		}
		return nil
	}
}

// Authenticated provides PreRun operations for commands that require a logged-in user
func (r *PreRun) Authenticated(cfg *config.Config, command *CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx, err := r.resolveContext(cfg, cmd, command, args)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		setClients(command, ctx, command.Version)
		err = r.resolveFlags(cfg, cmd, command)
		if err != nil {
			return err
		}
		// Validate token (not expired)
		state, err := ctx.AuthenticatedState()
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		var claims map[string]interface{}
		token, err := jwt.ParseSigned(state.AuthToken)
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
}

// HasAPIKey provides PreRun operations for commands that require an API key.
func (r *PreRun) HasAPIKey(cfg *config.Config, command *CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx, err := r.resolveContext(cfg, cmd, command, args)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		setClients(command, ctx, command.Version)
		err = r.resolveFlags(cfg, cmd, command)
		if err != nil {
			return err
		}
		clusterId := ctx.Kafka
		if clusterId == "" {
			return errors.HandleCommon(fmt.Errorf("context '%s' has no active Kafka cluster", ctx.Name), cmd)
		}
		if ctx.KafkaClusters[clusterId].APIKey == "" {
			err = &errors.UnspecifiedAPIKeyError{ClusterID: clusterId}
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

func (r *PreRun) resolveContext(cfg *config.Config, cmd *cobra.Command, cliCmd *CLICommand, args []string) (*config.Context, error) {
	err := r.Anonymous(cfg, cliCmd)(cmd, args)
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	ctx := cfg.Context()
	if ctx == nil {
		return nil, errors.HandleCommon(errors.ErrNoContext, cmd)
	}
	return ctx, nil
}

func setClients(cliCmd *CLICommand, ctx *config.Context, ver *version.Version) {
	cliCmd.Client = createCCloudClient(ctx, ver)
	cliCmd.MDSClient = createMDSClient(ctx, ver)
}

func (r *PreRun) resolveFlags(cfg *config.Config, cmd *cobra.Command, cliCmd *CLICommand) error {
	if err := r.FlagResolver.ResolveFlags(cmd, cfg, cliCmd.Client); err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func createCCloudClient(ctx *config.Context, ver *version.Version) *ccloud.Client {
	var baseURL string
	var authToken string
	var logger *log.Logger
	var userAgent string
	if ctx != nil {
		baseURL = ctx.Platform.Server
		state, err := ctx.AuthenticatedState()
		if err == nil {
			authToken = state.AuthToken
		}
		logger = ctx.Logger
		userAgent = ver.UserAgent
	}
	return ccloud.NewClientWithJWT(context.Background(), authToken, &ccloud.Params{
		BaseURL: baseURL, Logger: logger, UserAgent: userAgent,
	})
}

func createMDSClient(ctx *config.Context, ver *version.Version) *mds.APIClient {
	mdsConfig := mds.NewConfiguration()
	if ctx != nil {
		mdsConfig.BasePath = ctx.Platform.Server
		mdsConfig.UserAgent = ver.UserAgent
	}
	return mds.NewAPIClient(mdsConfig)
}
