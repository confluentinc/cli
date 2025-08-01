package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"

	pauth "github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/form"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/update"
	"github.com/confluentinc/cli/v4/pkg/utils"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
)

const autoLoginMsg = "Successful auto-login with non-interactive credentials."

// PreRun is a helper class for automatically setting up Cobra PersistentPreRun commands
type PreRunner interface {
	Anonymous(command *CLICommand, willAuthenticate bool) func(*cobra.Command, []string) error
	Authenticated(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error
	AuthenticatedWithMDS(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error
	InitializeOnPremKafkaRest(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error
	ParseFlagsIntoContext(command *CLICommand) func(*cobra.Command, []string) error
}

// PreRun is the standard PreRunner implementation
type PreRun struct {
	Config                  *config.Config
	Version                 *pversion.Version
	CCloudClientFactory     pauth.CCloudClientFactory
	MDSClientManager        pauth.MDSClientManager
	LoginCredentialsManager pauth.LoginCredentialsManager
	AuthTokenHandler        pauth.AuthTokenHandler
	JWTValidator            jwt.Validator
}

type KafkaRESTProvider func(cmd *cobra.Command) (*KafkaREST, error)

// Anonymous provides PreRun operations for commands that may be run without a logged-in user
func (r *PreRun) Anonymous(command *CLICommand, willAuthenticate bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Wait for a potential auto-login in the Authenticated PreRun function before checking run requirements.
		if !willAuthenticate {
			if err := ErrIfMissingRunRequirement(cmd, r.Config); err != nil {
				return err
			}
		}

		if err := r.Config.DecryptCredentials(); err != nil {
			return err
		}

		command.Config = r.Config
		if err := command.Config.ParseFlagsIntoConfig(cmd); err != nil {
			return err
		}

		// check Feature Flag "cli.disable" for commands run from cloud context (except for on-prem login)
		// check for commands that require cloud auth (since cloud context might not be active until auto-login)
		// check for cloud login (since it is not executed from cloud context)
		if (!isOnPremLoginCmd(command, r.Config.IsTest) && r.Config.IsCloudLogin()) || CommandRequiresCloudAuth(command.Command, command.Config) || isCloudLoginCmd(command, r.Config.IsTest) {
			if err := checkCliDisable(command, r.Config); err != nil {
				return err
			}
			// announcement and deprecation check, print out msg
			featureflags.PrintAnnouncements(r.Config, featureflags.Announcements, cmd)
			featureflags.PrintAnnouncements(r.Config, featureflags.DeprecationNotices, cmd)
		}

		// 1-4
		verbosity, err := cmd.Flags().GetCount("verbose")
		if err != nil {
			return err
		}

		// 5
		unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
		if err != nil {
			return err
		}
		if unsafeTrace {
			verbosity = int(log.UNSAFE_TRACE)
		}

		log.CliLogger.SetVerbosity(verbosity)
		log.CliLogger.Flush()

		command.Version = r.Version
		r.notifyIfUpdateAvailable(cmd, command.Version.Version)
		warnIfConfluentLocal(cmd)

		LabelRequiredFlags(cmd)

		return nil
	}
}

func checkCliDisable(cmd *CLICommand, cfg *config.Config) error {
	ldDisable := featureflags.GetLDDisableMap(cfg.Context())
	errMsg, errMsgOk := ldDisable["error_msg"].(string)
	disabledCmdsAndFlags, ok := ldDisable["patterns"].([]any)
	if (errMsgOk && errMsg != "" && !ok) || (ok && featureflags.IsDisabled(featureflags.Manager.Command, disabledCmdsAndFlags)) {
		allowUpdate, allowUpdateOk := ldDisable["allow_update"].(bool)
		if !(cmd.CommandPath() == "confluent update" && allowUpdateOk && allowUpdate) {
			// in case a user is trying to run an on-prem command from a cloud context (should not see LD msg)
			if err := ErrIfMissingRunRequirement(cmd.Command, cfg); err != nil && err == config.RequireOnPremLoginErr {
				return err
			}
			suggestionsMsg, _ := ldDisable["suggestions_msg"].(string)
			return errors.NewErrorWithSuggestions(errMsg, suggestionsMsg)
		}
	}
	return nil
}

func isOnPremLoginCmd(command *CLICommand, isTest bool) bool {
	if command.CommandPath() != "confluent login" {
		return false
	}
	mdsEnvUrl := os.Getenv(pauth.ConfluentPlatformMDSURL)
	url, _ := command.Flags().GetString("url")
	return (url == "" && mdsEnvUrl != "") || !ccloudv2.IsCCloudURL(url, isTest)
}

func isCloudLoginCmd(command *CLICommand, isTest bool) bool {
	if command.CommandPath() != "confluent login" {
		return false
	}
	mdsEnvUrl := os.Getenv(pauth.ConfluentPlatformMDSURL)
	url, _ := command.Flags().GetString("url")
	return (url == "" && mdsEnvUrl == "") || ccloudv2.IsCCloudURL(url, isTest)
}

func LabelRequiredFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if IsFlagRequired(flag) {
			flag.Usage = "REQUIRED: " + flag.Usage
		}
	})
}

func IsFlagRequired(flag *pflag.Flag) bool {
	required := flag.Annotations[cobra.BashCompOneRequiredFlag]
	if len(required) == 1 && required[0] == "true" {
		oneRequired := flag.Annotations["cobra_annotation_one_required"]
		if !(len(oneRequired) == 1 && slices.Contains(strings.Split(oneRequired[0], " "), flag.Name)) {
			return true
		}
	}
	return false
}

// Authenticated provides PreRun operations for commands that require a logged-in Confluent Cloud user.
func (r *PreRun) Authenticated(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := r.Anonymous(command.CLICommand, true)(cmd, args); err != nil {
			return err
		}

		if r.Config.Context().GetCredentialType() == config.APIKey {
			return ErrIfMissingRunRequirement(cmd, r.Config)
		}

		if err := r.Config.DecryptContextStates(); err != nil {
			return err
		}

		setContextErr := r.setAuthenticatedContext(command)
		if setContextErr != nil {
			if _, ok := setContextErr.(*errors.NotLoggedInError); ok {
				var machineName string
				if ctx := command.Config.Context(); ctx != nil {
					machineName = ctx.GetMachineName()
				}

				if err := r.ccloudAutoLogin(machineName); err != nil {
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

		unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
		if err != nil {
			return err
		}

		if tokenErr := r.ValidateToken(command.Config); tokenErr != nil {
			if err := r.updateToken(tokenErr, command.Config.Context(), unsafeTrace); err != nil {
				return err
			}
		}

		if err := r.setCCloudClient(command); err != nil {
			return err
		}

		command.V2Client = ccloudv2.NewClient(command.Config, unsafeTrace)

		return nil
	}
}

func (r *PreRun) ParseFlagsIntoContext(command *CLICommand) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return command.Config.Context().ParseFlagsIntoContext(cmd)
	}
}

func (r *PreRun) setAuthenticatedContext(cliCommand *AuthenticatedCLICommand) error {
	ctx := cliCommand.Config.Context()
	if !ctx.HasLogin() {
		return new(errors.NotLoggedInError)
	}
	cliCommand.Context = ctx

	return nil
}

func (r *PreRun) ccloudAutoLogin(machineName string) error {
	manager := pauth.NewLoginOrganizationManagerImpl()
	organizationId := pauth.GetLoginOrganization(
		manager.GetLoginOrganizationFromConfigurationFile(r.Config),
		manager.GetLoginOrganizationFromEnvironmentVariable(),
	)

	url := pauth.CCloudURL
	if ctxUrl := r.Config.Context().GetPlatformServer(); ctxUrl != "" {
		url = ctxUrl
	}

	credentials, err := r.getCCloudCredentials(machineName, url, organizationId)
	if err != nil {
		return err
	}

	if credentials == nil || credentials.AuthToken == "" {
		log.CliLogger.Debug("Non-interactive login failed: no credentials")
		return nil
	}

	client := r.CCloudClientFactory.JwtHTTPClientFactory(context.Background(), credentials.AuthToken, url)
	currentEnv, currentOrg, err := pauth.PersistCCloudCredentialsToConfig(r.Config, client, url, credentials, false)
	if err != nil {
		return err
	}

	log.CliLogger.Debug(autoLoginMsg)
	log.CliLogger.Debugf(errors.LoggedInAsMsgWithOrg, credentials.Username, currentOrg.ResourceId, currentOrg.Name)
	log.CliLogger.Debugf(errors.LoggedInUsingEnvMsg, currentEnv)

	return nil
}

func (r *PreRun) getCCloudCredentials(machineName, url, organizationId string) (*pauth.Credentials, error) {
	filterParams := config.MachineParams{
		Name:    machineName,
		IsCloud: true,
		URL:     url,
	}
	credentials, err := pauth.GetLoginCredentials(
		r.LoginCredentialsManager.GetCloudCredentialsFromEnvVar(organizationId),
		r.LoginCredentialsManager.GetCredentialsFromKeychain(true, filterParams.Name, url),
		r.LoginCredentialsManager.GetPrerunCredentialsFromConfig(r.Config),
		r.LoginCredentialsManager.GetCredentialsFromConfig(r.Config, filterParams),
	)
	if err != nil {
		log.CliLogger.Debugf("Auto-login failed to get credentials: %v", err)
		return nil, err
	}

	token, refreshToken, err := r.AuthTokenHandler.GetCCloudTokens(r.CCloudClientFactory, url, credentials, false, organizationId)
	if err != nil {
		return nil, err
	}
	credentials.AuthToken = token
	credentials.AuthRefreshToken = refreshToken

	return credentials, nil
}

func (r *PreRun) setCCloudClient(c *AuthenticatedCLICommand) error {
	c.Client = r.createCCloudClient(c.Context, c.Version)

	unsafeTrace, err := c.Flags().GetBool("unsafe-trace")
	if err != nil {
		return err
	}

	c.MDSv2Client = r.createMDSv2Client(c.Context, c.Version, unsafeTrace)

	provider := (KafkaRESTProvider)(func(cmd *cobra.Command) (*KafkaREST, error) {
		restEndpoint, lkc, err := getKafkaRestEndpoint(c.V2Client, c.Context)
		if err != nil {
			return nil, err
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil, err
		}
		cluster, httpResp, err := c.V2Client.DescribeKafkaCluster(lkc, environmentId)
		if err != nil {
			return nil, errors.CatchKafkaNotFoundError(err, lkc, httpResp)
		}
		if cluster.Status.Phase == ccloudv2.StatusProvisioning {
			return nil, fmt.Errorf(errors.KafkaRestProvisioningErrorMsg, lkc)
		}

		dataplaneToken, err := pauth.GetDataplaneToken(c.Context)
		if err != nil {
			return nil, err
		}

		if restEndpoint == "" {
			return nil, fmt.Errorf("Kafka REST is not enabled: the operation is only supported with Kafka REST proxy")
		}

		activeEndpoint := c.Context.KafkaClusterContext.GetActiveKafkaClusterEndpoint()

		// sanity check of whether the cluster specified in command corresponds to the endpoint specified
		endpointMatchCluster := false

		clusterEndpoints := cluster.Spec.GetEndpoints()

		for _, attributes := range clusterEndpoints {
			if attributes.GetHttpEndpoint() == activeEndpoint {
				endpointMatchCluster = true
				break
			}
		}

		// stored config precedes default value
		// set restEndpoint to value stored in CLI config if present
		if activeEndpoint != "" && endpointMatchCluster {
			restEndpoint = activeEndpoint
		}

		flagEndpoint, err := cmd.Flags().GetString("kafka-endpoint")
		if err != nil {
			return nil, err
		}

		// input flag precedes stored config value
		// if the endpoint flag is set, use its value; otherwise, use the value from config.RestEndpoint
		if flagEndpoint != "" {
			restEndpoint = flagEndpoint
		}

		kafkaRest := &KafkaREST{
			Context:     context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, dataplaneToken),
			CloudClient: ccloudv2.NewKafkaRestClient(restEndpoint, lkc, r.Version.UserAgent, dataplaneToken, unsafeTrace),
			Client:      CreateKafkaRESTClient(restEndpoint, unsafeTrace),
		}
		return kafkaRest, nil
	})
	c.KafkaRESTProvider = &provider
	return nil
}

func getKafkaRestEndpoint(client *ccloudv2.Client, ctx *config.Context) (string, string, error) {
	config, err := kafka.GetClusterForCommand(client, ctx)
	if err != nil {
		return "", "", err
	}

	return config.RestEndpoint, config.ID, err
}

func (r *PreRun) createCCloudClient(ctx *config.Context, ver *pversion.Version) *ccloudv1.Client {
	params := &ccloudv1.Params{
		BaseURL:   ctx.GetPlatformServer(),
		Logger:    log.CliLogger,
		UserAgent: ver.UserAgent,
	}

	return ccloudv1.NewClientWithJWT(context.Background(), ctx.GetAuthToken(), params)
}

// Authenticated provides PreRun operations for commands that require a logged-in MDS user.
func (r *PreRun) AuthenticatedWithMDS(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := r.Anonymous(command.CLICommand, true)(cmd, args); err != nil {
			return err
		}

		if err := r.Config.DecryptContextStates(); err != nil {
			return err
		}

		setContextErr := r.setAuthenticatedContext(command)
		if setContextErr != nil {
			if _, ok := setContextErr.(*errors.NotLoggedInError); ok {
				if err := r.confluentAutoLogin(cmd); err != nil {
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

		unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
		if err != nil {
			return err
		}

		if tokenErr := r.ValidateToken(command.Config); tokenErr != nil {
			if err := r.updateToken(tokenErr, command.Config.Context(), unsafeTrace); err != nil {
				return err
			}
		}

		return nil
	}
}

func (r *PreRun) confluentAutoLogin(cmd *cobra.Command) error {
	token, refreshToken, credentials, err := r.getConfluentTokenAndCredentials(cmd)
	if err != nil {
		return err
	}
	if token == "" || credentials == nil {
		log.CliLogger.Debug("Non-interactive login failed: no credentials")
		return nil
	}
	if err := pauth.PersistConfluentLoginToConfig(r.Config, credentials, credentials.PrerunLoginURL, token, refreshToken, credentials.PrerunLoginCaCertPath, false); err != nil {
		return err
	}
	log.CliLogger.Debug(autoLoginMsg)
	log.CliLogger.Debugf(errors.LoggedInAsMsg, credentials.Username)
	return nil
}

func (r *PreRun) getConfluentTokenAndCredentials(cmd *cobra.Command) (string, string, *pauth.Credentials, error) {
	if pauth.IsOnPremSSOEnv() {
		// Skip auto-login when CONFLUENT_PLATFORM_SSO=true
		return "", "", nil, nil
	}

	credentials, err := pauth.GetLoginCredentials(
		r.LoginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar(),
	)
	if err != nil {
		return "", "", nil, err
	}

	unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
	if err != nil {
		return "", "", nil, err
	}

	client, err := r.MDSClientManager.GetMDSClient(credentials.PrerunLoginURL, credentials.PrerunLoginCaCertPath, "", "", unsafeTrace)
	if err != nil {
		return "", "", nil, err
	}
	token, refreshToken, err := r.AuthTokenHandler.GetConfluentToken(client, credentials, false)
	if err != nil {
		return "", "", nil, err
	}

	return token, refreshToken, credentials, err
}

// InitializeOnPremKafkaRest provides PreRun operations for on-prem commands that require a Kafka REST Proxy client. (ccloud RP commands use Authenticated prerun)
// Initializes a default KafkaRestClient
func (r *PreRun) InitializeOnPremKafkaRest(command *AuthenticatedCLICommand) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// pass mds token as bearer token otherwise use http basic auth
		// no error means user is logged in with mds and has valid token; on an error we try http basic auth since mds is not needed for RP commands
		err := r.AuthenticatedWithMDS(command)(cmd, args)
		if _, ok := err.(*errors.RunRequirementError); ok {
			return err
		}
		useMdsToken := err == nil

		provider := (KafkaRESTProvider)(func(_ *cobra.Command) (*KafkaREST, error) {
			cfg := kafkarestv3.NewConfiguration()

			unsafeTrace, err := cmd.Flags().GetBool("unsafe-trace")
			if err != nil {
				return nil, err
			}
			cfg.Debug = unsafeTrace

			restFlags, err := resolveOnPremKafkaRestFlags(cmd)
			if err != nil {
				return nil, err
			}
			cfg.HTTPClient, err = createOnPremKafkaRestClient(command.Context, restFlags.caCertPath, restFlags.clientCertPath, restFlags.clientKeyPath, log.CliLogger)
			if err != nil {
				return nil, err
			}
			client := kafkarestv3.NewAPIClient(cfg)
			if restFlags.noAuth {
				return &KafkaREST{
					Client:  client,
					Context: context.Background(),
				}, nil
			}
			var restContext context.Context
			if useMdsToken && !restFlags.prompt {
				log.CliLogger.Debug("found mds token to use as bearer")
				restContext = context.WithValue(context.Background(), kafkarestv3.ContextAccessToken, command.Context.GetAuthToken())
			} else if restFlags.clientCertPath != "" && !restFlags.prompt { // credentials not needed for mTLS auth
				restContext = context.Background()
			} else { // no mds token, then prompt for basic auth creds
				if !restFlags.prompt {
					output.Println(r.Config.EnableColor, "No session token found, please enter user credentials. To avoid being prompted, run `confluent login`.")
				}
				f := form.New(
					form.Field{ID: "username", Prompt: "Username"},
					form.Field{ID: "password", Prompt: "Password", IsHidden: true},
				)
				if err := f.Prompt(form.NewPrompt()); err != nil {
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

type onPremKafkaRestFlagValues struct {
	url            string
	caCertPath     string
	clientCertPath string
	clientKeyPath  string
	noAuth         bool
	prompt         bool
}

func resolveOnPremKafkaRestFlags(cmd *cobra.Command) (*onPremKafkaRestFlagValues, error) {
	url, _ := cmd.Flags().GetString("url")
	certificateAAuthorityPath, _ := cmd.Flags().GetString("certificate-authority-path")
	clientCertPath, clientKeyPath, _ := GetClientCertAndKeyPaths(cmd)
	noAuthentication, _ := cmd.Flags().GetBool("no-authentication")
	prompt, _ := cmd.Flags().GetBool("prompt")

	if (clientCertPath == "") != (clientKeyPath == "") {
		return nil, fmt.Errorf(errors.NeedClientCertAndKeyPathsErrorMsg)
	}

	values := &onPremKafkaRestFlagValues{
		url:            url,
		caCertPath:     certificateAAuthorityPath,
		clientCertPath: clientCertPath,
		clientKeyPath:  clientKeyPath,
		noAuth:         noAuthentication,
		prompt:         prompt,
	}

	return values, nil
}

func createOnPremKafkaRestClient(ctx *config.Context, caCertPath, clientCertPath, clientKeyPath string, logger *log.Logger) (*http.Client, error) {
	if caCertPath == "" {
		caCertPath = os.Getenv(pauth.ConfluentPlatformCertificateAuthorityPath)
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
	} else if ctx != nil && ctx.Platform != nil && ctx.Platform.CaCertPath != "" { // if no cert-path flag is specified, use the cert path from the config
		client, err := utils.CustomCAAndClientCertClient(ctx.Platform.CaCertPath, clientCertPath, clientKeyPath)
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

func (r *PreRun) ValidateToken(config *config.Config) error {
	if config == nil {
		return new(errors.NotLoggedInError)
	}
	ctx := config.Context()
	if ctx == nil {
		return new(errors.NotLoggedInError)
	}
	return r.JWTValidator.Validate(ctx)
}

func (r *PreRun) updateToken(tokenErr error, ctx *config.Context, unsafeTrace bool) error {
	log.CliLogger.Debug("Updating auth tokens")
	token, refreshToken, err := r.getUpdatedAuthToken(ctx, unsafeTrace)
	if err != nil || token == "" {
		log.CliLogger.Debug("Failed to update auth tokens")
		_ = ctx.DeleteUserAuth()

		if _, ok := tokenErr.(*ccloudv1.InvalidTokenError); ok {
			tokenErr = new(ccloudv1.InvalidTokenError)
		}

		return tokenErr
	}

	log.CliLogger.Debug("Successfully updated auth tokens")
	return ctx.UpdateAuthTokens(token, refreshToken)
}

func (r *PreRun) getUpdatedAuthToken(ctx *config.Context, unsafeTrace bool) (string, string, error) {
	filterParams := config.MachineParams{
		IsCloud: r.Config.IsCloudLogin(),
		Name:    ctx.GetMachineName(),
	}

	if r.Config.IsCloudLogin() {
		manager := pauth.NewLoginOrganizationManagerImpl()
		organizationId := pauth.GetLoginOrganization(
			manager.GetLoginOrganizationFromConfigurationFile(r.Config),
			manager.GetLoginOrganizationFromEnvironmentVariable(),
		)

		credentials, err := pauth.GetLoginCredentials(
			r.LoginCredentialsManager.GetCloudCredentialsFromEnvVar(organizationId),
			r.LoginCredentialsManager.GetCredentialsFromKeychain(true, ctx.Name, ctx.GetPlatformServer()),
			r.LoginCredentialsManager.GetPrerunCredentialsFromConfig(r.Config),
			r.LoginCredentialsManager.GetCredentialsFromConfig(r.Config, filterParams),
		)
		if err != nil {
			return "", "", err
		}

		return r.AuthTokenHandler.GetCCloudTokens(r.CCloudClientFactory, ctx.GetPlatformServer(), credentials, false, organizationId)
	} else {
		credentials, err := pauth.GetLoginCredentials(
			r.LoginCredentialsManager.GetOnPremCredentialsFromEnvVar(),
			r.LoginCredentialsManager.GetCredentialsFromKeychain(false, ctx.Name, ctx.GetPlatformServer()),
			r.LoginCredentialsManager.GetPrerunCredentialsFromConfig(r.Config),
			r.LoginCredentialsManager.GetCredentialsFromConfig(r.Config, filterParams),
			r.LoginCredentialsManager.GetOnPremSsoCredentialsFromConfig(r.Config, unsafeTrace),
		)
		if err != nil {
			return "", "", err
		}

		mdsClientManager := pauth.MDSClientManagerImpl{}
		client, err := mdsClientManager.GetMDSClient(ctx.GetPlatformServer(), ctx.Platform.CaCertPath, "", "", unsafeTrace)
		if err != nil {
			return "", "", err
		}
		return r.AuthTokenHandler.GetConfluentToken(client, credentials, false)
	}
}

// notifyIfUpdateAvailable prints a message if an update is available
func (r *PreRun) notifyIfUpdateAvailable(cmd *cobra.Command, _ string) {
	if !r.shouldCheckForUpdates(cmd) {
		return
	}

	current, err := version.NewVersion(r.Config.Version.Version)
	if err != nil {
		return
	}

	client := update.NewClient(r.Config.IsTest)

	binaries, err := client.GetBinaries()
	if err != nil {
		return
	}

	minorVersions, majorVersions := update.FilterUpdates(binaries, current)

	if len(majorVersions) > 0 {
		output.ErrPrintf(r.Config.EnableColor, "A major version update is available for %s from (current: %s, latest: %s).\n", pversion.CLIName, current, majorVersions[len(majorVersions)-1])
		output.ErrPrintf(r.Config.EnableColor, "To view release notes and install the update, please run `confluent update --major`.\n")
		output.ErrPrintf(r.Config.EnableColor, "\n")
	}

	if len(minorVersions) > 0 {
		output.ErrPrintf(r.Config.EnableColor, "A minor version update is available for %s from (current: %s, latest: %s).\n", pversion.CLIName, current, minorVersions[len(minorVersions)-1])
		output.ErrPrintf(r.Config.EnableColor, "To view release notes and install the update, please run `confluent update`.\n")
		output.ErrPrintf(r.Config.EnableColor, "\n")
	}
}

func (r *PreRun) shouldCheckForUpdates(cmd *cobra.Command) bool {
	if r.Config.IsTest || r.Config.DisableUpdates || r.Config.DisableUpdateCheck {
		return false
	}

	for _, subcommand := range []string{"prompt", "update"} {
		if strings.HasPrefix(cmd.CommandPath(), fmt.Sprintf("confluent %s", subcommand)) {
			return false
		}
	}

	// Only check for updates once a day
	if r.Config.LastUpdateCheckAt != nil && time.Since(*r.Config.LastUpdateCheckAt) < 24*time.Hour {
		return false
	}

	now := time.Now()
	r.Config.LastUpdateCheckAt = &now
	_ = r.Config.Save()

	return true
}

func warnIfConfluentLocal(cmd *cobra.Command) {
	if strings.HasPrefix(cmd.CommandPath(), "confluent local kafka start") {
		output.ErrPrintln(false, "The local commands are intended for a single-node development environment only, NOT for production usage. See more: https://docs.confluent.io/current/cli/index.html")
		output.ErrPrintln(false, "")
		return
	}
	if strings.HasPrefix(cmd.CommandPath(), "confluent local") && !strings.HasPrefix(cmd.CommandPath(), "confluent local kafka") {
		output.ErrPrintln(false, "The local commands are intended for a single-node development environment only, NOT for production usage. See more: https://docs.confluent.io/current/cli/index.html")
		output.ErrPrintln(false, "As of Confluent Platform 8.0, Java 8 will no longer be supported.")
		output.ErrPrintln(false, "")
	}
}

func (r *PreRun) createMDSv2Client(ctx *config.Context, ver *pversion.Version, unsafeTrace bool) *mdsv2alpha1.APIClient {
	mdsv2Config := mdsv2alpha1.NewConfiguration()
	mdsv2Config.HTTPClient = utils.DefaultClient()
	mdsv2Config.Debug = unsafeTrace
	if ctx == nil {
		return mdsv2alpha1.NewAPIClient(mdsv2Config)
	}
	mdsv2Config.BasePath = ctx.GetPlatformServer() + "/api/metadata/security/v2alpha1"
	mdsv2Config.UserAgent = ver.UserAgent

	return mdsv2alpha1.NewAPIClient(mdsv2Config)
}

func CreateKafkaRESTClient(kafkaRestURL string, unsafeTrace bool) *kafkarestv3.APIClient {
	cfg := kafkarestv3.NewConfiguration()
	cfg.HTTPClient = utils.DefaultClient()
	cfg.Debug = unsafeTrace
	cfg.BasePath = kafkaRestURL + "/kafka/v3"
	return kafkarestv3.NewAPIClient(cfg)
}
