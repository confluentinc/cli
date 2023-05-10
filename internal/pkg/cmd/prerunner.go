package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/update"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

// PreRun is a helper class for automatically setting up Cobra PersistentPreRun commands
type PreRunner interface {
	Anonymous(command *CLICommand, willAuthenticate bool) func(*cobra.Command, []string) error
	Authenticated(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error
	AuthenticatedWithMDS(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error
	HasAPIKey(command *HasAPIKeyCLICommand) func(*cobra.Command, []string) error
	InitializeOnPremKafkaRest(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error
	ParseFlagsIntoContext(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error
	AnonymousParseFlagsIntoContext(command *CLICommand) func(*cobra.Command, []string) error
}

// PreRun is the standard PreRunner implementation
type PreRun struct {
	Config                  *v1.Config
	UpdateClient            update.Client
	FlagResolver            FlagResolver
	Version                 *version.Version
	CCloudClientFactory     pauth.CCloudClientFactory
	MDSClientManager        pauth.MDSClientManager
	LoginCredentialsManager pauth.LoginCredentialsManager
	AuthTokenHandler        pauth.AuthTokenHandler
	JWTValidator            JWTValidator
	IsTest                  bool
}

type CLICommand struct {
	*cobra.Command
	Config    *DynamicConfig
	Version   *version.Version
	prerunner PreRunner
}

type KafkaRESTProvider func() (*KafkaREST, error)

type AuthenticatedCLICommand struct {
	*CLICommand
	Client            *ccloud.Client
	MDSClient         *mds.APIClient
	MDSv2Client       *mdsv2alpha1.APIClient
	KafkaRESTProvider *KafkaRESTProvider
	Context           *DynamicContext
	State             *v1.ContextState
}

type AuthenticatedStateFlagCommand struct {
	*AuthenticatedCLICommand
}

type StateFlagCommand struct {
	*CLICommand
}

type HasAPIKeyCLICommand struct {
	*CLICommand
	Context *DynamicContext
}

func NewAuthenticatedCLICommand(cmd *cobra.Command, prerunner PreRunner) *AuthenticatedCLICommand {
	c := &AuthenticatedCLICommand{CLICommand: NewCLICommand(cmd, prerunner)}
	cmd.PersistentPreRunE = NewCLIPreRunnerE(prerunner.Authenticated(c))
	c.Command = cmd
	return c
}

func (c *AuthenticatedCLICommand) SetPersistentPreRunE(persistentPreRunE func(*cobra.Command, []string) error) {
	c.PersistentPreRunE = NewCLIPreRunnerE(persistentPreRunE)
}

// Returns AuthenticatedStateFlagCommand used for cloud authenticated commands that require (or have child commands that require) state flags (i.e. cluster, environment, context)
func NewAuthenticatedStateFlagCommand(cmd *cobra.Command, prerunner PreRunner) *AuthenticatedStateFlagCommand {
	c := &AuthenticatedStateFlagCommand{NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.PersistentPreRunE = NewCLIPreRunnerE(prerunner.Authenticated(c.AuthenticatedCLICommand), prerunner.ParseFlagsIntoContext(c.AuthenticatedCLICommand))
	c.Command = cmd
	return c
}

// Returns AuthenticatedStateFlagCommand used for mds authenticated commands that require (or have child commands that require) state flags (i.e. context)
func NewAuthenticatedWithMDSStateFlagCommand(cmd *cobra.Command, prerunner PreRunner) *AuthenticatedStateFlagCommand {
	c := &AuthenticatedStateFlagCommand{NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}
	cmd.PersistentPreRunE = NewCLIPreRunnerE(prerunner.AuthenticatedWithMDS(c.AuthenticatedCLICommand), prerunner.ParseFlagsIntoContext(c.AuthenticatedCLICommand))
	c.Command = cmd
	return c
}

// Returns StateFlagCommand used for non-authenticated commands that require (or have child commands that require) state flags (i.e. cluster, environment, context)
func NewAnonymousStateFlagCommand(cmd *cobra.Command, prerunner PreRunner) *StateFlagCommand {
	c := &StateFlagCommand{NewAnonymousCLICommand(cmd, prerunner)}
	cmd.PersistentPreRunE = NewCLIPreRunnerE(prerunner.Anonymous(c.CLICommand, false), prerunner.AnonymousParseFlagsIntoContext(c.CLICommand))
	c.Command = cmd
	return c
}

func NewAuthenticatedWithMDSCLICommand(cmd *cobra.Command, prerunner PreRunner) *AuthenticatedCLICommand {
	c := &AuthenticatedCLICommand{CLICommand: NewCLICommand(cmd, prerunner)}
	cmd.PersistentPreRunE = NewCLIPreRunnerE(prerunner.AuthenticatedWithMDS(c))
	c.Command = cmd
	return c
}

func NewHasAPIKeyCLICommand(cmd *cobra.Command, prerunner PreRunner) *HasAPIKeyCLICommand {
	c := &HasAPIKeyCLICommand{CLICommand: NewCLICommand(cmd, prerunner)}
	cmd.PersistentPreRunE = NewCLIPreRunnerE(prerunner.HasAPIKey(c))
	c.Command = cmd
	return c
}

func NewAnonymousCLICommand(cmd *cobra.Command, prerunner PreRunner) *CLICommand {
	c := NewCLICommand(cmd, prerunner)
	cmd.PersistentPreRunE = NewCLIPreRunnerE(prerunner.Anonymous(c, false))
	c.Command = cmd
	return c
}

func NewCLICommand(command *cobra.Command, prerunner PreRunner) *CLICommand {
	return &CLICommand{
		Config:    &DynamicConfig{},
		Command:   command,
		prerunner: prerunner,
	}
}

func (s *AuthenticatedStateFlagCommand) AddCommand(command *cobra.Command) {
	s.AuthenticatedCLICommand.AddCommand(command)
}

func (c *AuthenticatedCLICommand) AddCommand(command *cobra.Command) {
	c.Command.AddCommand(command)
}

func (s *StateFlagCommand) AddCommand(command *cobra.Command) {
	s.Command.AddCommand(command)
}

func (c *AuthenticatedCLICommand) GetKafkaREST() (*KafkaREST, error) {
	return (*c.KafkaRESTProvider)()
}

func (c *AuthenticatedCLICommand) AuthToken() string {
	return c.State.AuthToken
}

func (c *AuthenticatedCLICommand) EnvironmentId() string {
	return c.State.Auth.Account.Id
}

func (h *HasAPIKeyCLICommand) AddCommand(command *cobra.Command) {
	command.PersistentPreRunE = h.PersistentPreRunE
	h.Command.AddCommand(command)
}

// Anonymous provides PreRun operations for commands that may be run without a logged-in user
func (r *PreRun) Anonymous(command *CLICommand, willAuthenticate bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Wait for a potential auto-login in the Authenticated PreRun function before checking run requirements.
		if !willAuthenticate {
			if err := ErrIfMissingRunRequirement(cmd, r.Config); err != nil {
				return err
			}
		}

		if err := command.Config.InitDynamicConfig(cmd, r.Config, r.FlagResolver); err != nil {
			return err
		}
		if err := log.SetLoggingVerbosity(cmd, log.CliLogger); err != nil {
			return err
		}
		log.CliLogger.Flush()
		command.Version = r.Version
		r.notifyIfUpdateAvailable(cmd, command.Version.Version)
		r.warnIfConfluentLocal(cmd)

		if r.Config != nil {
			ctx := command.Config.Context()
			err := r.ValidateToken(cmd, command.Config)
			switch err.(type) {
			case *ccloud.ExpiredTokenError:
				if err := ctx.DeleteUserAuth(); err != nil {
					return err
				}
				utils.ErrPrintln(cmd, errors.TokenExpiredMsg)
			}
		}

		LabelRequiredFlags(cmd)

		return nil
	}
}

func LabelRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if IsFlagRequired(flag) {
			flag.Usage = "REQUIRED: " + flag.Usage
		}
	})
}

func IsFlagRequired(flag *pflag.Flag) bool {
	annotations := flag.Annotations[cobra.BashCompOneRequiredFlag]
	return len(annotations) == 1 && annotations[0] == "true"
}

// Authenticated provides PreRun operations for commands that require a logged-in Confluent Cloud user.
func (r *PreRun) Authenticated(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := r.Anonymous(command.CLICommand, true)(cmd, args); err != nil {
			return err
		}

		setContextErr := r.setAuthenticatedContext(command)
		if setContextErr != nil {
			if _, ok := setContextErr.(*errors.NotLoggedInError); ok { //nolint:gosimple // false positive
				var netrcMachineName string
				if ctx := command.Config.Context(); ctx != nil {
					netrcMachineName = ctx.GetNetrcMachineName()
				}

				if err := r.ccloudAutoLogin(cmd, netrcMachineName); err != nil {
					log.CliLogger.Debugf("Auto login failed: %v", err)
				} else {
					setContextErr = r.setAuthenticatedContext(command)
				}
			} else {
				return setContextErr
			}
		}

		// Even if there was an error while setting the context, notify the user about any unmet run requirements first.
		if err := ErrIfMissingRunRequirement(cmd, r.Config); err != nil {
			return err
		}

		if setContextErr != nil {
			return setContextErr
		}

		if err := r.ValidateToken(cmd, command.Config); err != nil {
			return err
		}
		return r.setCCloudClient(command)
	}
}

func (r *PreRun) ParseFlagsIntoContext(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := command.Context
		return ctx.ParseFlagsIntoContext(cmd, command.Client)
	}
}

func (r *PreRun) AnonymousParseFlagsIntoContext(command *CLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return command.Config.Context().ParseFlagsIntoContext(cmd, nil)
	}
}

func (r *PreRun) setAuthenticatedContext(cliCommand *AuthenticatedCLICommand) error {
	ctx := cliCommand.Config.Context()
	if ctx == nil {
		return new(errors.NotLoggedInError)
	}
	cliCommand.Context = ctx

	state, err := ctx.AuthenticatedState()
	if err != nil {
		return err
	}
	cliCommand.State = state

	return nil
}

func (r *PreRun) ccloudAutoLogin(cmd *cobra.Command, netrcMachineName string) error {
	orgResourceId := r.Config.GetLastUsedOrgId()
	token, credentials, err := r.getCCloudTokenAndCredentials(cmd, netrcMachineName, orgResourceId)
	if err != nil {
		return err
	}
	if token == "" || credentials == nil {
		log.CliLogger.Debug("Non-interactive login failed: no credentials")
		return nil
	}

	client := r.CCloudClientFactory.JwtHTTPClientFactory(context.Background(), credentials.AuthToken, pauth.CCloudURL)
	currentEnv, currentOrg, err := pauth.PersistCCloudLoginToConfig(r.Config, credentials, pauth.CCloudURL, client, false)
	if err != nil {
		return err
	}
	log.CliLogger.Debug(errors.AutoLoginMsg)
	log.CliLogger.Debugf(errors.LoggedInAsMsgWithOrg, credentials.Username, currentOrg.ResourceId, currentOrg.Name)
	log.CliLogger.Debugf(errors.LoggedInUsingEnvMsg, currentEnv.Id, currentEnv.Name)
	return nil
}

func (r *PreRun) getCCloudTokenAndCredentials(cmd *cobra.Command, netrcMachineName, orgResourceId string) (string, *pauth.Credentials, error) {
	filterParams := netrc.NetrcMachineParams{
		Name:    netrcMachineName,
		IsCloud: true,
		URL:     pauth.CCloudURL,
	}

	credentials, err := pauth.GetLoginCredentials(
		r.LoginCredentialsManager.GetCloudCredentialsFromEnvVar(cmd, orgResourceId),
		r.LoginCredentialsManager.GetCredentialsFromKeychain(r.Config, true, filterParams.Name, pauth.CCloudURL),
		r.LoginCredentialsManager.GetPrerunCredentialsFromConfig(r.Config, filterParams),
		r.LoginCredentialsManager.GetCredentialsFromNetrc(cmd, filterParams),
		r.LoginCredentialsManager.GetCredentialsFromConfig(r.Config, filterParams),
	)
	if err != nil {
		log.CliLogger.Debugf("Prerun login getting credentials failed: %v", err.Error())
		return "", nil, err
	}

	token, _, err := r.AuthTokenHandler.GetCCloudTokens(r.CCloudClientFactory, pauth.CCloudURL, credentials, false, orgResourceId)
	if err != nil {
		return "", nil, err
	}

	return token, credentials, err
}

func (r *PreRun) setCCloudClient(cliCmd *AuthenticatedCLICommand) error {
	ctx := cliCmd.Config.Context()

	ccloudClient, err := r.createCCloudClient(ctx, cliCmd.Version)
	if err != nil {
		return err
	}
	cliCmd.Client = ccloudClient
	cliCmd.Context.client = ccloudClient
	cliCmd.Config.Client = ccloudClient
	cliCmd.MDSv2Client = r.createMDSv2Client(ctx, cliCmd.Version)
	provider := (KafkaRESTProvider)(func() (*KafkaREST, error) {
		ctx := cliCmd.Config.Context()

		restEndpoint, lkc, err := getKafkaRestEndpoint(ctx, cliCmd)
		if err != nil {
			return nil, err
		}
		if restEndpoint != "" {
			result := &KafkaREST{}
			result.Client, err = createKafkaRESTClient(restEndpoint)
			if err != nil {
				return nil, err
			}
			state, err := ctx.AuthenticatedState()
			if err != nil {
				return nil, err
			}
			bearerToken, err := pauth.GetBearerToken(state, ctx.Platform.Server, lkc)
			if err != nil {
				return nil, err
			}
			result.Context = context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, bearerToken)
			return result, nil
		}
		return nil, nil
	})
	cliCmd.KafkaRESTProvider = &provider
	return nil
}

func getKafkaRestEndpoint(ctx *DynamicContext, cmd *AuthenticatedCLICommand) (string, string, error) {
	if os.Getenv("XX_CCLOUD_USE_KAFKA_API") != "" {
		return "", "", nil
	}
	clusterConfig, err := ctx.GetKafkaClusterForCommand()
	if err != nil {
		return "", "", err
	}
	if clusterConfig.RestEndpoint != "" {
		return clusterConfig.RestEndpoint, clusterConfig.ID, nil
	}
	// if clusterConfig.RestEndpoint is empty, fetch the cluster to ensure config isn't just out of date
	// potentially remove this once Rest Proxy is enabled across prod
	client := NewContextClient(ctx)
	kafkaCluster, err := client.FetchCluster(clusterConfig.ID)
	if err != nil {
		return "", clusterConfig.ID, err
	}
	// no need to update the config if it's still empty
	if kafkaCluster.RestEndpoint == "" {
		return "", clusterConfig.ID, nil
	}
	// update config to have updated cluster if rest endpoint is no longer ""
	refreshedClusterConfig := KafkaClusterToKafkaClusterConfig(kafkaCluster)
	ctx.KafkaClusterContext.AddKafkaClusterConfig(refreshedClusterConfig)
	err = ctx.Save() //should we fail on this error or log and continue?
	if err != nil {
		return "", clusterConfig.ID, err
	}
	return kafkaCluster.RestEndpoint, clusterConfig.ID, nil
}

// Converts a ccloud base URL to the appropriate Metrics URL.
func ConvertToMetricsBaseURL(baseURL string) string {
	// strip trailing slashes before comparing.
	trimmedURL := strings.TrimRight(baseURL, "/")
	if trimmedURL == "https://confluent.cloud" {
		return "https://api.telemetry.confluent.cloud/"
	} else if strings.HasSuffix(trimmedURL, "priv.cpdev.cloud") || trimmedURL == "https://devel.cpdev.cloud" {
		return "https://devel-sandbox-api.telemetry.aws.confluent.cloud/"
	} else if trimmedURL == "https://stag.cpdev.cloud" {
		return "https://stag-sandbox-api.telemetry.aws.confluent.cloud/"
	}
	// if no matches, then use original URL
	return baseURL
}

func (r *PreRun) createCCloudClient(ctx *DynamicContext, ver *version.Version) (*ccloud.Client, error) {
	var baseURL string
	var authToken string
	var userAgent string
	if ctx != nil {
		baseURL = ctx.Platform.Server
		state, err := ctx.AuthenticatedState()
		if err != nil {
			return nil, err
		}
		authToken = state.AuthToken
		userAgent = ver.UserAgent
	}
	return ccloud.NewClientWithJWT(context.Background(), authToken, &ccloud.Params{
		BaseURL: baseURL, Logger: log.CliLogger, UserAgent: userAgent, MetricsBaseURL: ConvertToMetricsBaseURL(baseURL),
	}), nil
}

// Authenticated provides PreRun operations for commands that require a logged-in MDS user.
func (r *PreRun) AuthenticatedWithMDS(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := r.Anonymous(command.CLICommand, true)(cmd, args); err != nil {
			return err
		}

		setContextErr := r.setAuthenticatedWithMDSContext(command)
		if setContextErr != nil {
			if _, ok := setContextErr.(*errors.NotLoggedInError); ok { //nolint:gosimple // false positive
				var netrcMachineName string
				if ctx := command.Config.Context(); ctx != nil {
					netrcMachineName = ctx.GetNetrcMachineName()
				}

				if err := r.confluentAutoLogin(cmd, netrcMachineName); err != nil {
					log.CliLogger.Debugf("Auto login failed: %v", err)
				} else {
					setContextErr = r.setAuthenticatedWithMDSContext(command)
				}
			} else {
				return setContextErr
			}
		}

		// Even if there was an error while setting the context, notify the user about any unmet run requirements first.
		if err := ErrIfMissingRunRequirement(cmd, r.Config); err != nil {
			return err
		}

		if setContextErr != nil {
			return setContextErr
		}

		return r.ValidateToken(cmd, command.Config)
	}
}

func (r *PreRun) setAuthenticatedWithMDSContext(cliCommand *AuthenticatedCLICommand) error {
	ctx := cliCommand.Config.Context()
	if ctx == nil || !ctx.HasBasicMDSLogin() {
		return new(errors.NotLoggedInError)
	}
	cliCommand.Context = ctx
	cliCommand.State = ctx.State
	r.setConfluentClient(cliCommand)
	return nil
}

func (r *PreRun) confluentAutoLogin(cmd *cobra.Command, netrcMachineName string) error {
	token, credentials, err := r.getConfluentTokenAndCredentials(cmd, netrcMachineName)
	if err != nil {
		return err
	}
	if token == "" || credentials == nil {
		log.CliLogger.Debug("Non-interactive login failed: no credentials")
		return nil
	}
	err = pauth.PersistConfluentLoginToConfig(r.Config, credentials, credentials.PrerunLoginURL, token, credentials.PrerunLoginCaCertPath, false, false)
	if err != nil {
		return err
	}
	log.CliLogger.Debug(errors.AutoLoginMsg)
	log.CliLogger.Debugf(errors.LoggedInAsMsg, credentials.Username)
	return nil
}

func (r *PreRun) getConfluentTokenAndCredentials(cmd *cobra.Command, netrcMachineName string) (string, *pauth.Credentials, error) {
	filterParams := netrc.NetrcMachineParams{
		Name:    netrcMachineName,
		IsCloud: false,
	}

	credentials, err := pauth.GetLoginCredentials(
		r.LoginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar(cmd),
		r.LoginCredentialsManager.GetOnPremPrerunCredentialsFromNetrc(cmd, filterParams),
	)
	if err != nil {
		return "", nil, err
	}

	client, err := r.MDSClientManager.GetMDSClient(credentials.PrerunLoginURL, credentials.PrerunLoginCaCertPath)
	if err != nil {
		return "", nil, err
	}
	token, err := r.AuthTokenHandler.GetConfluentToken(client, credentials)
	if err != nil {
		return "", nil, err
	}

	return token, credentials, err
}

func (r *PreRun) setConfluentClient(cliCmd *AuthenticatedCLICommand) {
	ctx := cliCmd.Config.Context()
	cliCmd.MDSClient = r.createMDSClient(ctx, cliCmd.Version)
}

func (r *PreRun) createMDSClient(ctx *DynamicContext, ver *version.Version) *mds.APIClient {
	mdsConfig := mds.NewConfiguration()
	mdsConfig.HTTPClient = utils.DefaultClient()
	if log.CliLogger.GetLevel() >= log.DEBUG {
		mdsConfig.Debug = true
	}
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
	client, err := utils.SelfSignedCertClientFromPath(caCertPath)
	if err != nil {
		log.CliLogger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --ca-cert-path flag.", caCertPath, err.Error())
	} else {
		mdsConfig.HTTPClient = client
	}
	return mds.NewAPIClient(mdsConfig)
}

// InitializeOnPremKafkaRest provides PreRun operations for on-prem commands that require a Kafka REST Proxy client. (ccloud RP commands use Authenticated prerun)
// Initializes a default KafkaRestClient
func (r *PreRun) InitializeOnPremKafkaRest(command *AuthenticatedCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// pass mds token as bearer token otherwise use http basic auth
		// no error means user is logged in with mds and has valid token; on an error we try http basic auth since mds is not needed for RP commands
		err := r.AuthenticatedWithMDS(command)(cmd, args)
		useMdsToken := err == nil

		provider := (KafkaRESTProvider)(func() (*KafkaREST, error) {
			cfg := kafkarestv3.NewConfiguration()
			restFlags, err := r.FlagResolver.ResolveOnPremKafkaRestFlags(cmd)
			if err != nil {
				return nil, err
			}
			cfg.HTTPClient, err = createOnPremKafkaRestClient(command.Context, restFlags.caCertPath, restFlags.clientCertPath, restFlags.clientKeyPath, log.CliLogger)
			if err != nil {
				return nil, err
			}
			client := kafkarestv3.NewAPIClient(cfg)
			if restFlags.noAuth || restFlags.clientCertPath != "" { // credentials not needed for mTLS auth
				return &KafkaREST{
					Client:  client,
					Context: context.Background(),
				}, nil
			}
			var restContext context.Context
			if useMdsToken && !restFlags.prompt {
				log.CliLogger.Debug("found mds token to use as bearer")
				restContext = context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, command.AuthToken())
			} else { // no mds token, then prompt for basic auth creds
				if !restFlags.prompt {
					utils.Println(cmd, errors.MDSTokenNotFoundMsg)
				}
				f := form.New(
					form.Field{ID: "username", Prompt: "Username"},
					form.Field{ID: "password", Prompt: "Password", IsHidden: true},
				)
				if err := f.Prompt(command.Command, form.NewPrompt(os.Stdin)); err != nil {
					return nil, err
				}
				restContext = context.WithValue(context.Background(), kafkarestv3.ContextBasicAuth, kafkarestv3.BasicAuth{UserName: f.Responses["username"].(string), Password: f.Responses["password"].(string)})
			}
			return &KafkaREST{
				Client:  client,
				Context: restContext,
			}, nil
		})
		command.KafkaRESTProvider = &provider
		return nil
	}
}

func createOnPremKafkaRestClient(ctx *DynamicContext, caCertPath string, clientCertPath string, clientKeyPath string, logger *log.Logger) (*http.Client, error) {
	if caCertPath == "" {
		caCertPath = pauth.GetEnvWithFallback(pauth.ConfluentPlatformCACertPath, pauth.DeprecatedConfluentPlatformCACertPath)
		logger.Debugf("Found CA cert path: %s", caCertPath)
	}
	// use cert path flag or env var if it was passed
	if caCertPath != "" {
		client, err := utils.CustomCAAndClientCertClient(caCertPath, clientCertPath, clientKeyPath)
		if err != nil {
			return nil, err
		}
		return client, nil
		// use cert path from config if available
	} else if ctx != nil && ctx.Context != nil && ctx.Context.Platform != nil && ctx.Context.Platform.CaCertPath != "" { //if no cert-path flag is specified, use the cert path from the config
		client, err := utils.CustomCAAndClientCertClient(ctx.Context.Platform.CaCertPath, clientCertPath, clientKeyPath)
		if err != nil {
			return nil, err
		}
		return client, nil
	} else if clientCertPath != "" && clientKeyPath != "" {
		client, err := utils.CustomCAAndClientCertClient("", clientCertPath, clientKeyPath)
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return utils.DefaultClient(), nil
}

// HasAPIKey provides PreRun operations for commands that require an API key.
func (r *PreRun) HasAPIKey(command *HasAPIKeyCLICommand) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := r.Anonymous(command.CLICommand, false)(cmd, args); err != nil {
			return err
		}

		ctx := command.Config.Context()
		if ctx == nil {
			return new(errors.NotLoggedInError)
		}
		command.Context = ctx

		var clusterId string
		if command.Context.Credential.CredentialType == v1.APIKey {
			clusterId = r.getClusterIdForAPIKeyCredential(ctx)
		} else if command.Context.Credential.CredentialType == v1.Username {
			if err := r.ValidateToken(cmd, command.Config); err != nil {
				return err
			}

			client, err := r.createCCloudClient(ctx, command.Version)
			if err != nil {
				return err
			}
			ctx.client = client
			command.Config.Client = client

			if err := ctx.ParseFlagsIntoContext(cmd, command.Config.Client); err != nil {
				return err
			}

			cluster, err := ctx.GetKafkaClusterForCommand()
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
					cluster.APIKeys[key] = &v1.APIKeyPair{Key: key, Secret: secret}
				} else if cluster.APIKeys[key] == nil {
					return errors.NewErrorWithSuggestions(
						fmt.Sprintf(errors.NoAPISecretStoredOrPassedMsg, key, clusterId),
						fmt.Sprintf(errors.NoAPISecretStoredOrPassedSuggestions, key, clusterId))
				}
			}
		} else {
			panic("Invalid Credential Type")
		}

		hasAPIKey, err := ctx.HasAPIKey(clusterId)
		if err != nil {
			return err
		}
		if !hasAPIKey {
			return &errors.UnspecifiedAPIKeyError{ClusterID: clusterId}
		}

		return nil
	}
}

func (r *PreRun) ValidateToken(cmd *cobra.Command, config *DynamicConfig) error {
	if config == nil {
		return new(errors.NotLoggedInError)
	}
	ctx := config.Context()
	if ctx == nil {
		return new(errors.NotLoggedInError)
	}
	err := r.JWTValidator.Validate(ctx.Context)
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
		log.CliLogger.Debug("Dynamic context is nil. Cannot attempt to update auth token.")
		return tokenError
	}
	log.CliLogger.Debug("Updating auth token")
	token, err := r.getUpdatedAuthToken(cmd, ctx)
	if err != nil || token == "" {
		log.CliLogger.Debug("Failed to update auth token")
		return tokenError
	}
	log.CliLogger.Debug("Successfully update auth token")
	err = ctx.UpdateAuthToken(token)
	if err != nil {
		return tokenError
	}
	return nil
}

func (r *PreRun) getUpdatedAuthToken(cmd *cobra.Command, ctx *DynamicContext) (string, error) {
	filterParams := netrc.NetrcMachineParams{
		IsCloud: r.Config.IsCloudLogin(),
		Name:    ctx.GetNetrcMachineName(),
	}
	credentials, err := pauth.GetLoginCredentials(
		r.LoginCredentialsManager.GetPrerunCredentialsFromConfig(ctx.Config, filterParams),
		r.LoginCredentialsManager.GetCredentialsFromNetrc(cmd, filterParams),
		r.LoginCredentialsManager.GetCredentialsFromConfig(r.Config, filterParams),
	)
	if err != nil {
		return "", err
	}

	var token string
	if r.Config.IsCloudLogin() {
		orgResourceId := r.Config.GetLastUsedOrgId()
		token, _, err = r.AuthTokenHandler.GetCCloudTokens(r.CCloudClientFactory, ctx.Platform.Server, credentials, false, orgResourceId)
		if err != nil {
			return "", err
		}
	} else {
		mdsClientManager := pauth.MDSClientManagerImpl{}
		client, err := mdsClientManager.GetMDSClient(ctx.Platform.Server, ctx.Platform.CaCertPath)
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

// if API key credential then the context is initialized to be used for only one cluster, and cluster id can be obtained directly from the context config
func (r *PreRun) getClusterIdForAPIKeyCredential(ctx *DynamicContext) string {
	return ctx.KafkaClusterContext.GetActiveKafkaClusterId()
}

// notifyIfUpdateAvailable prints a message if an update is available
func (r *PreRun) notifyIfUpdateAvailable(cmd *cobra.Command, currentVersion string) {
	if !r.shouldCheckForUpdates(cmd) || r.IsTest {
		return
	}

	latestMajorVersion, latestMinorVersion, err := r.UpdateClient.CheckForUpdates(version.CLIName, currentVersion, false)
	if err != nil {
		// This is a convenience helper to check-for-updates before arbitrary commands. Since the CLI supports running
		// in internet-less environments (e.g., local or on-prem deploys), swallow the error and log a warning.
		log.CliLogger.Warn(err)
		return
	}

	if latestMajorVersion != "" {
		if !strings.HasPrefix(latestMajorVersion, "v") {
			latestMajorVersion = "v" + latestMajorVersion
		}
		utils.ErrPrintf(cmd, errors.NotifyMajorUpdateMsg, version.CLIName, currentVersion, latestMajorVersion, version.CLIName)
	}

	if latestMinorVersion != "" {
		if !strings.HasPrefix(latestMinorVersion, "v") {
			latestMinorVersion = "v" + latestMinorVersion
		}
		utils.ErrPrintf(cmd, errors.NotifyMinorUpdateMsg, version.CLIName, currentVersion, latestMinorVersion, version.CLIName)
	}
}

func (r *PreRun) shouldCheckForUpdates(cmd *cobra.Command) bool {
	for _, subcommand := range []string{"prompt", "update"} {
		if strings.HasPrefix(cmd.CommandPath(), fmt.Sprintf("confluent %s", subcommand)) {
			return false
		}
	}

	return true
}

func (r *PreRun) warnIfConfluentLocal(cmd *cobra.Command) {
	if strings.HasPrefix(cmd.CommandPath(), "confluent local") {
		//nolint
		utils.ErrPrintln(cmd, errors.LocalCommandDevOnlyMsg)
	}
}

func (r *PreRun) createMDSv2Client(ctx *DynamicContext, ver *version.Version) *mdsv2alpha1.APIClient {
	mdsv2Config := mdsv2alpha1.NewConfiguration()
	mdsv2Config.HTTPClient = utils.DefaultClient()
	if log.CliLogger.GetLevel() >= log.DEBUG {
		mdsv2Config.Debug = true
	}
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
	client, err := utils.SelfSignedCertClientFromPath(caCertPath)
	if err != nil {
		log.CliLogger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --ca-cert-path flag.", caCertPath, err.Error())
	} else {
		mdsv2Config.HTTPClient = client
	}
	return mdsv2alpha1.NewAPIClient(mdsv2Config)
}

func createKafkaRESTClient(kafkaRestURL string) (*kafkarestv3.APIClient, error) {
	cfg := kafkarestv3.NewConfiguration()
	cfg.HTTPClient = utils.DefaultClient()
	if log.CliLogger.GetLevel() >= log.DEBUG {
		cfg.Debug = true
	}
	cfg.BasePath = kafkaRestURL + "/kafka/v3"
	return kafkarestv3.NewAPIClient(cfg), nil
}
