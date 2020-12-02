package cmd

import (
	"context"
	"os"
	"strings"

	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"

	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/update"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// PreRun is a helper class for automatically setting up Cobra PersistentPreRun commands
type PreRunner interface {
	Anonymous(command *CLICommand) func(cmd *cobra.Command, args []string) error
	Authenticated(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error
	AuthenticatedWithMDS(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error
	HasAPIKey(command *HasAPIKeyCLICommand) func(cmd *cobra.Command, args []string) error
}

const DoNotTrack = "do-not-track-analytics"

// PreRun is the standard PreRunner implementation
type PreRun struct {
	Config                  *v3.Config
	ConfigLoadingError      error
	UpdateClient            update.Client
	CLIName                 string
	Logger                  *log.Logger
	Analytics               analytics.Client
	FlagResolver            FlagResolver
	Version                 *version.Version
	LoginCredentialsManager pauth.LoginCredentialsManager
	AuthTokenHandler        pauth.AuthTokenHandler
	JWTValidator            JWTValidator
}

type CLICommand struct {
	*cobra.Command
	Config    *DynamicConfig
	Version   *version.Version
	prerunner PreRunner
}

type AuthenticatedCLICommand struct {
	*CLICommand
	Client      *ccloud.Client
	MDSClient   *mds.APIClient
	MDSv2Client *mdsv2alpha1.APIClient
	Context     *DynamicContext
	State       *v2.ContextState
}

type AuthenticatedStateFlagCommand struct {
	*AuthenticatedCLICommand
	subcommandFlags map[string]*pflag.FlagSet
}

type HasAPIKeyCLICommand struct {
	*CLICommand
	Context         *DynamicContext
	subcommandFlags map[string]*pflag.FlagSet
}

func (r *PreRun) ValidateToken(cmd *cobra.Command, config *DynamicConfig) error {
	if config == nil {
		return &errors.NoContextError{CLIName: r.CLIName}
	}
	ctx, err := config.Context(cmd)
	if err != nil {
		return err
	}
	if ctx == nil {
		return &errors.NoContextError{CLIName: r.CLIName}
	}
	err = r.JWTValidator.Validate(ctx.Context)
	if err == nil {
		return nil
	}
	switch err.(type) {
	case *ccloud.InvalidTokenError:
		return r.updateToken(new(ccloud.InvalidTokenError), cmd, ctx)
	case *ccloud.ExpiredTokenError:
		return r.updateToken(new(ccloud.ExpiredTokenError), cmd, ctx)
	}
	if err.Error() == errors.MalformedJWTNoExprErrorMsg {
		return r.updateToken(errors.New(errors.MalformedJWTNoExprErrorMsg), cmd, ctx)
	} else {
		return r.updateToken(err, cmd, ctx)
	}
}

func (r *PreRun) updateToken(tokenError error, cmd *cobra.Command, ctx *DynamicContext) error {
	if ctx == nil {
		r.Logger.Debug("Dynamic context is nil. Cannot attempt to update auth token.")
		return tokenError
	}
	r.Logger.Debug("Updating auth token")
	token, err := r.getNewAuthToken(cmd, ctx)
	if err != nil || token == "" {
		r.Logger.Debug("Failed to update auth token")
		if err != nil {
			r.Logger.Debugf("Update token error : %s", err.Error())
		}
		return tokenError
	}
	r.Logger.Debug("Successfully updated auth token")
	err = ctx.UpdateAuthToken(token)
	if err != nil {
		return tokenError
	}
	return nil
}

func (r *PreRun) getNewAuthToken(cmd *cobra.Command, ctx *DynamicContext) (string, error) {
	params := netrc.GetMatchingNetrcMachineParams{
		CLIName: r.CLIName,
		CtxName: ctx.Name,
	}
	credentials, err := pauth.GetLoginCredentials(r.LoginCredentialsManager.GetCredentialsFromNetrc(cmd, params))
	if err != nil {
		return "", err
	}

	var token string
	if r.CLIName == "ccloud" {
		client := ccloud.NewClient(&ccloud.Params{BaseURL: ctx.Platform.Server, HttpClient: ccloud.BaseClient, Logger: r.Logger, UserAgent: r.Version.UserAgent})
		token, _, err = r.AuthTokenHandler.GetCCloudTokens(client, credentials, false)
		if err != nil {
			return "", err
		}
	} else {
		mdsClientManager := pauth.MDSClientManagerImpl{}
		client, err := mdsClientManager.GetMDSClient(ctx.Platform.Server, ctx.Platform.CaCertPath, r.Logger)
		if err != nil {
			return "", err
		}
		token, err = r.AuthTokenHandler.GetConfluentToken(client, credentials)
		if err != nil {
			return "", err
		}
	}
	return token, nil
}

func (a *AuthenticatedCLICommand) AuthToken() string {
	return a.State.AuthToken
}

func (a *AuthenticatedCLICommand) EnvironmentId() string {
	return a.State.Auth.Account.Id
}

func NewAuthenticatedCLICommand(command *cobra.Command, prerunner PreRunner) *AuthenticatedCLICommand {
	cmd := &AuthenticatedCLICommand{
		CLICommand: NewCLICommand(command, prerunner),
		Context:    nil,
		State:      nil,
	}
	command.PersistentPreRunE = NewCLIPreRunnerE(prerunner.Authenticated(cmd))
	cmd.Command = command
	return cmd
}

func NewAuthenticatedStateFlagCommand(command *cobra.Command, prerunner PreRunner, flagMap map[string]*pflag.FlagSet) *AuthenticatedStateFlagCommand {
	cmd := &AuthenticatedStateFlagCommand{
		NewAuthenticatedCLICommand(command, prerunner),
		flagMap,
	}
	return cmd
}

func NewAuthenticatedWithMDSCLICommand(command *cobra.Command, prerunner PreRunner) *AuthenticatedCLICommand {
	cmd := &AuthenticatedCLICommand{
		CLICommand: NewCLICommand(command, prerunner),
		Context:    nil,
		State:      nil,
	}
	command.PersistentPreRunE = NewCLIPreRunnerE(prerunner.AuthenticatedWithMDS(cmd))
	cmd.Command = command
	return cmd
}

func NewHasAPIKeyCLICommand(command *cobra.Command, prerunner PreRunner, flagMap map[string]*pflag.FlagSet) *HasAPIKeyCLICommand {
	cmd := &HasAPIKeyCLICommand{
		CLICommand:      NewCLICommand(command, prerunner),
		Context:         nil,
		subcommandFlags: flagMap,
	}
	command.PersistentPreRunE = NewCLIPreRunnerE(prerunner.HasAPIKey(cmd))
	cmd.Command = command
	return cmd
}

func NewAnonymousCLICommand(command *cobra.Command, prerunner PreRunner) *CLICommand {
	cmd := NewCLICommand(command, prerunner)
	command.PersistentPreRunE = NewCLIPreRunnerE(prerunner.Anonymous(cmd))
	cmd.Command = command
	return cmd
}

func NewCLICommand(command *cobra.Command, prerunner PreRunner) *CLICommand {
	return &CLICommand{
		Config:    &DynamicConfig{},
		Command:   command,
		prerunner: prerunner,
	}
}

func (s *AuthenticatedStateFlagCommand) AddCommand(command *cobra.Command) {
	command.Flags().AddFlagSet(s.subcommandFlags[strings.Fields(s.Use)[0]])
	command.Flags().AddFlagSet(s.subcommandFlags[strings.Fields(command.Use)[0]])
	command.Flags().SortFlags = false
	s.AuthenticatedCLICommand.AddCommand(command)
}

func (a *AuthenticatedCLICommand) AddCommand(command *cobra.Command) {
	command.PersistentPreRunE = a.PersistentPreRunE
	a.Command.AddCommand(command)
}

func (h *HasAPIKeyCLICommand) AddCommand(command *cobra.Command) {
	command.Flags().AddFlagSet(h.subcommandFlags[strings.Fields(h.Use)[0]])
	command.Flags().AddFlagSet(h.subcommandFlags[strings.Fields(command.Use)[0]])
	command.PersistentPreRunE = h.PersistentPreRunE
	h.Command.AddCommand(command)
}

// CanCompleteCommand returns whether or not the specified command can be completed.
// If the prerunner of the command returns no error, true is returned,
// and if an error is encountered, false is returned.
func CanCompleteCommand(cmd *cobra.Command) bool {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations[DoNotTrack] = ""
	err := cmd.PersistentPreRunE(cmd, []string{})
	delete(cmd.Annotations, DoNotTrack)
	return err == nil
}

// Anonymous provides PreRun operations for commands that may be run without a logged-in user
func (r *PreRun) Anonymous(command *CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if _, ok := cmd.Annotations[DoNotTrack]; !ok {
			r.Analytics.TrackCommand(cmd, args)
		}
		err := command.Config.InitDynamicConfig(cmd, r.Config, r.FlagResolver)
		if err != nil {
			return err
		}
		command.Version = r.Version
		if err := log.SetLoggingVerbosity(cmd, r.Logger); err != nil {
			return err
		}
		r.Logger.Flush()
		if err := r.notifyIfUpdateAvailable(cmd, r.CLIName, command.Version.Version); err != nil {
			return err
		}
		r.warnIfConfluentLocal(cmd)
		if r.Config != nil {
			ctx, err := command.Config.Context(cmd)
			if err != nil {
				return err
			}
			err = r.ValidateToken(cmd, command.Config)
			switch err.(type) {
			case *ccloud.ExpiredTokenError:
				err := ctx.DeleteUserAuth()
				if err != nil {
					return err
				}
				utils.ErrPrintln(cmd, errors.TokenExpiredMsg)
				analyticsError := r.Analytics.SessionTimedOut()
				if analyticsError != nil {
					r.Logger.Debug(analyticsError.Error())
				}
			}
		} else {
			if isAuthOrConfigCommands(cmd) {
				return r.ConfigLoadingError
			}
		}
		LabelRequiredFlags(cmd)
		return nil
	}
}

func isAuthOrConfigCommands(cmd *cobra.Command) bool {
	return strings.Contains(cmd.CommandPath(), "login") ||
		strings.Contains(cmd.CommandPath(), "logout") ||
		strings.Contains(cmd.CommandPath(), "config")
}

func LabelRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		annotations := flag.Annotations[cobra.BashCompOneRequiredFlag]
		if len(annotations) == 1 && annotations[0] == "true" {
			flag.Usage = "REQUIRED: " + flag.Usage
		}
	})
}

// Authenticated provides PreRun operations for commands that require a logged-in Confluent Cloud user.
func (r *PreRun) Authenticated(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := r.Anonymous(command.CLICommand)(cmd, args)
		if err != nil {
			return err
		}
		if r.Config == nil {
			return r.ConfigLoadingError
		}
		err = r.setCCloudClient(command)
		if err != nil {
			return err
		}
		ctx, err := command.Config.Context(cmd)
		if err != nil {
			return err
		}
		if ctx == nil {
			return &errors.NoContextError{CLIName: r.CLIName}
		}
		command.Context = ctx
		command.State, err = ctx.AuthenticatedState(cmd)
		if err != nil {
			return err
		}
		return r.ValidateToken(cmd, command.Config)
	}
}

func (r *PreRun) setCCloudClient(cliCmd *AuthenticatedCLICommand) error {
	ctx, err := cliCmd.Config.Context(cliCmd.Command)
	if err != nil {
		return err
	}
	ccloudClient, err := r.createCCloudClient(ctx, cliCmd.Command, cliCmd.Version)
	if err != nil {
		return err
	}
	cliCmd.Client = ccloudClient
	cliCmd.Config.Client = ccloudClient
	cliCmd.MDSv2Client = r.createMDSv2Client(ctx, cliCmd.Version)
	return nil
}

func (r *PreRun) createCCloudClient(ctx *DynamicContext, cmd *cobra.Command, ver *version.Version) (*ccloud.Client, error) {
	var baseURL string
	var authToken string
	var logger *log.Logger
	var userAgent string
	if ctx != nil {
		baseURL = ctx.Platform.Server
		state, err := ctx.AuthenticatedState(cmd)
		if err != nil {
			return nil, err
		}
		authToken = state.AuthToken
		logger = ctx.Logger
		userAgent = ver.UserAgent
	}
	return ccloud.NewClientWithJWT(context.Background(), authToken, &ccloud.Params{
		BaseURL: baseURL, Logger: logger, UserAgent: userAgent,
	}), nil
}

// Authenticated provides PreRun operations for commands that require a logged-in MDS user.
func (r *PreRun) AuthenticatedWithMDS(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := r.Anonymous(command.CLICommand)(cmd, args)
		if err != nil {
			return err
		}
		if r.Config == nil {
			return r.ConfigLoadingError
		}
		err = r.setConfluentClient(command)
		if err != nil {
			return err
		}
		ctx, err := command.Config.Context(cmd)
		if err != nil {
			return err
		}
		if ctx == nil {
			return &errors.NoContextError{CLIName: r.CLIName}
		}
		if !ctx.HasMDSLogin() {
			return &errors.NotLoggedInError{CLIName: r.CLIName}
		}
		command.Context = ctx
		command.State = ctx.State
		return r.ValidateToken(cmd, command.Config)
	}
}

func (r *PreRun) setConfluentClient(cliCmd *AuthenticatedCLICommand) error {
	ctx, err := cliCmd.Config.Context(cliCmd.Command)
	if err != nil {
		return err
	}
	cliCmd.MDSClient = r.createMDSClient(ctx, cliCmd.Version)
	return nil
}

func (r *PreRun) createMDSClient(ctx *DynamicContext, ver *version.Version) *mds.APIClient {
	mdsConfig := mds.NewConfiguration()
	if ctx == nil {
		return mds.NewAPIClient(mdsConfig)
	}
	mdsConfig.BasePath = ctx.Platform.Server
	mdsConfig.UserAgent = ver.UserAgent
	if ctx.Platform.CaCertPath == "" {
		return mds.NewAPIClient(mdsConfig)
	}
	caCertPath := ctx.Platform.CaCertPath
	// Try to load certs. On failure, warn, but don't error out because this may be an auth command, so there may
	// be a --ca-cert-path flag on the cmd line that'll fix whatever issue there is with the cert file in the config
	caCertFile, err := os.Open(caCertPath)
	if err == nil {
		defer caCertFile.Close()
		mdsConfig.HTTPClient, err = pauth.SelfSignedCertClient(caCertFile, r.Logger)
		if err != nil {
			r.Logger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --ca-cert-path flag.", caCertPath, err.Error())
			mdsConfig.HTTPClient = pauth.DefaultClient()
		}
	} else {
		r.Logger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --ca-cert-path flag.", caCertPath, err.Error())
		mdsConfig.HTTPClient = pauth.DefaultClient()

	}
	return mds.NewAPIClient(mdsConfig)
}

// HasAPIKey provides PreRun operations for commands that require an API key.
func (r *PreRun) HasAPIKey(command *HasAPIKeyCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := r.Anonymous(command.CLICommand)(cmd, args)
		if err != nil {
			return err
		}
		if r.Config == nil {
			return r.ConfigLoadingError
		}
		ctx, err := command.Config.Context(cmd)
		if err != nil {
			return err
		}
		if ctx == nil {
			return &errors.NoContextError{CLIName: r.CLIName}
		}
		command.Context = ctx
		var clusterId string
		if command.Context.Credential.CredentialType == v2.APIKey {
			clusterId = r.getClusterIdForAPIKeyCredential(ctx)
		} else if command.Context.Credential.CredentialType == v2.Username {
			err := r.ValidateToken(cmd, command.Config)
			if err != nil {
				return err
			}
			client, err := r.createCCloudClient(ctx, cmd, command.Version)
			if err != nil {
				return err
			}
			ctx.client = client
			command.Config.Client = client
			cluster, err := ctx.GetKafkaClusterForCommand(cmd)
			if err != nil {
				return err
			}
			clusterId = cluster.ID
			key, secret, err := ctx.KeyAndSecretFlags(cmd)
			if err != nil {
				return err
			}
			if key != "" {
				cluster.APIKey = key
				if secret != "" {
					cluster.APIKeys[key] = &v0.APIKeyPair{Key: key, Secret: secret}
				} else if cluster.APIKeys[key] == nil {
					return errors.NewErrorWithSuggestions(errors.NoAPISecretStoredOrPassedMsg, errors.NoAPISecretStoredOrPassedSuggestions)
				}
			}
		} else {
			panic("Invalid Credential Type")
		}
		hasAPIKey, err := ctx.HasAPIKey(cmd, clusterId)
		if err != nil {
			return err
		}
		if !hasAPIKey {
			err = &errors.UnspecifiedAPIKeyError{ClusterID: clusterId}
			return err
		}
		return nil
	}
}

// if API key credential then the context is initialized to be used for only one cluster, and cluster id can be obtained directly from the context config
func (r *PreRun) getClusterIdForAPIKeyCredential(ctx *DynamicContext) string {
	return ctx.KafkaClusterContext.GetActiveKafkaClusterId()
}

// notifyIfUpdateAvailable prints a message if an update is available
func (r *PreRun) notifyIfUpdateAvailable(cmd *cobra.Command, name string, currentVersion string) error {
	if isUpdateCommand(cmd) {
		return nil
	}
	updateAvailable, latestVersion, err := r.UpdateClient.CheckForUpdates(name, currentVersion, false)
	if err != nil {
		// This is a convenience helper to check-for-updates before arbitrary commands. Since the CLI supports running
		// in internet-less environments (e.g., local or on-prem deploys), swallow the error and log a warning.
		r.Logger.Warn(err)
		return nil
	}
	if updateAvailable {
		if !strings.HasPrefix(latestVersion, "v") {
			latestVersion = "v" + latestVersion
		}
		utils.ErrPrintf(cmd, errors.NotifyUpdateMsg, name, currentVersion, latestVersion, name)
	}
	return nil
}

func isUpdateCommand(cmd *cobra.Command) bool {
	return strings.Contains(cmd.CommandPath(), "update")
}

func (r *PreRun) warnIfConfluentLocal(cmd *cobra.Command) {
	if strings.HasPrefix(cmd.CommandPath(), "confluent local") {
		utils.ErrPrintln(cmd, errors.LocalCommandDevOnlyMsg)
	}
}

func (r *PreRun) createMDSv2Client(ctx *DynamicContext, ver *version.Version) *mdsv2alpha1.APIClient {
	mdsv2Config := mdsv2alpha1.NewConfiguration()
	if ctx == nil {
		return mdsv2alpha1.NewAPIClient(mdsv2Config)
	}
	mdsv2Config.BasePath = ctx.Platform.Server + "/api/metadata/security/v2alpha1"
	mdsv2Config.UserAgent = ver.UserAgent
	if ctx.Platform.CaCertPath == "" {
		return mdsv2alpha1.NewAPIClient(mdsv2Config)
	}
	caCertPath := ctx.Platform.CaCertPath
	// Try to load certs. On failure, warn, but don't error out because this may be an auth command, so there may
	// be a --ca-cert-path flag on the cmd line that'll fix whatever issue there is with the cert file in the config
	caCertFile, err := os.Open(caCertPath)
	if err == nil {
		defer caCertFile.Close()
		mdsv2Config.HTTPClient, err = pauth.SelfSignedCertClient(caCertFile, r.Logger)
		if err != nil {
			r.Logger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --ca-cert-path flag.", caCertPath, err.Error())
			mdsv2Config.HTTPClient = pauth.DefaultClient()
		}
	} else {
		r.Logger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --ca-cert-path flag.", caCertPath, err.Error())
		mdsv2Config.HTTPClient = pauth.DefaultClient()

	}
	return mdsv2alpha1.NewAPIClient(mdsv2Config)
}
