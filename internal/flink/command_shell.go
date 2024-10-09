package flink

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	client "github.com/confluentinc/cli/v4/pkg/flink/app"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/jwt"
	"github.com/confluentinc/cli/v4/pkg/log"
	ppanic "github.com/confluentinc/cli/v4/pkg/panic-recovery"
)

func (c *command) newShellCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "shell",
		Short:       "Start Flink interactive SQL client.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.startFlinkSqlClient(prerunner, cmd)
		},
	}

	c.addComputePoolFlag(cmd)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	c.addDatabaseFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	if featureflags.Manager.BoolVariation("cli.flink.internal", c.Context, config.CliLaunchDarklyClient, true, false) {
		cmd.Flags().StringSlice("config-key", []string{}, "App option keys for local mode.")
		cmd.Flags().StringSlice("config-value", []string{}, "App option values for local mode.")
	}

	return cmd
}

func (c *command) authenticated(authenticated func(*cobra.Command, []string) error, cmd *cobra.Command, jwtValidator jwt.Validator) func() error {
	return func() error {
		authToken := c.Context.GetAuthToken()
		authRefreshToken := c.Context.GetAuthRefreshToken()
		if err := c.Context.UpdateAuthTokens(authToken, authRefreshToken); err != nil {
			return err
		}

		if err := authenticated(cmd, nil); err != nil {
			return err
		}

		flinkGatewayClient, err := c.GetFlinkGatewayClient(true)
		if err != nil {
			return err
		}

		jwtCtx := &config.Context{State: &config.ContextState{AuthToken: flinkGatewayClient.AuthToken}}
		if tokenErr := jwtValidator.Validate(jwtCtx); tokenErr != nil {
			dataplaneToken, err := auth.GetDataplaneToken(c.Context)
			if err != nil {
				return err
			}
			flinkGatewayClient.AuthToken = dataplaneToken
		}

		return nil
	}
}

func (c *command) startFlinkSqlClient(prerunner pcmd.PreRunner, cmd *cobra.Command) error {
	if featureflags.Manager.BoolVariation("cli.flink.internal", c.Context, config.CliLaunchDarklyClient, true, false) {
		// get config keys and values from flags
		configKeys, err := cmd.Flags().GetStringSlice("config-key")
		if err != nil {
			return err
		}
		configValues, err := cmd.Flags().GetStringSlice("config-value")
		if err != nil {
			return err
		}

		// if configs were passed, we should enter local mode
		if len(configKeys) > 0 && len(configValues) > 0 {
			return c.startWithLocalMode(configKeys, configValues)
		}
	}

	environmentId, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environmentId == "" {
		if c.Context.GetCurrentEnvironment() == "" {
			return errors.NewErrorWithSuggestions(
				"no environment provided",
				"Provide an environment with `confluent environment use env-123456` or `--environment`.",
			)
		}
		environmentId = c.Context.GetCurrentEnvironment()
	}

	catalog := c.Context.GetCurrentFlinkCatalog()
	if catalog == "" {
		environment, err := c.V2Client.GetOrgEnvironment(environmentId)
		if err != nil {
			return errors.NewErrorWithSuggestions(err.Error(), "List available environments with `confluent environment list`.")
		}
		catalog = environment.GetDisplayName()
	}

	computePool := c.Context.GetCurrentFlinkComputePool()
	if computePool == "" {
		return errors.NewErrorWithSuggestions(
			"no compute pool selected",
			"Select a compute pool with `confluent flink compute-pool use` or `--compute-pool`.",
		)
	}

	serviceAccount, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}
	if serviceAccount == "" {
		serviceAccount = c.Context.GetCurrentServiceAccount()
	}

	database, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}
	if database == "" {
		if c.Context.GetCurrentFlinkDatabase() != "" {
			database = c.Context.GetCurrentFlinkDatabase()
		} else {
			database = c.Context.KafkaClusterContext.GetActiveKafkaClusterConfig().GetName()
		}
	}

	unsafeTrace, err := c.Command.Flags().GetBool("unsafe-trace")
	if err != nil {
		return err
	}

	flinkGatewayClient, err := c.GetFlinkGatewayClient(true)
	if err != nil {
		return err
	}

	lspBaseUrl, err := c.getFlinkLanguageServiceUrl(flinkGatewayClient)
	if err != nil {
		log.CliLogger.Warnf("Flink shell failed to connect to language service: error getting language service URL: %v\n", err)
		return err
	}

	jwtValidator := jwt.NewValidator()

	verbose, _ := cmd.Flags().GetCount("verbose")

	opts := types.ApplicationOptions{
		Context:          c.Context,
		UnsafeTrace:      unsafeTrace,
		UserAgent:        c.Version.UserAgent,
		EnvironmentName:  catalog,
		EnvironmentId:    environmentId,
		OrganizationId:   c.Context.GetOrganization().GetResourceId(),
		Database:         database,
		ComputePoolId:    computePool,
		ServiceAccountId: serviceAccount,
		Verbose:          verbose > 0,
		LSPBaseUrl:       lspBaseUrl,
	}

	return client.StartApp(flinkGatewayClient, c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd, jwtValidator), opts, reportUsage(cmd, c.Config, unsafeTrace))
}

func (c *command) startWithLocalMode(configKeys, configValues []string) error {
	// parse app options from given flags
	appOptions, err := types.ParseApplicationOptionsFromSlices(configKeys, configValues)
	if err != nil {
		return err
	}

	// validate app options
	if err := appOptions.Validate(); err != nil {
		return err
	}

	gatewayClient := ccloudv2.NewFlinkGatewayClient(appOptions.GetGatewayUrl(), c.Version.UserAgent, appOptions.GetUnsafeTrace(), "authToken")

	appOptions.Context = c.Context
	return client.StartApp(gatewayClient, func() error { return nil }, *appOptions, func() {})
}

func (c *command) getFlinkLanguageServiceUrl(gatewayClient *ccloudv2.FlinkGatewayClient) (string, error) {
	if cfg := gatewayClient.GetConfig(); cfg != nil && len(cfg.Servers) > 0 {
		gatewayUrl := cfg.Servers[0].URL
		parsedUrl, err := url.Parse(gatewayUrl)
		if err != nil {
			return "", err
		}

		parsedUrl.Host = strings.Replace(parsedUrl.Host, "flink.", "flinkpls.", 1)
		parsedUrl.Scheme = "wss"
		parsedUrl.Path = "/lsp"

		return parsedUrl.String(), nil
	}
	return "", nil
}

func reportUsage(cmd *cobra.Command, cfg *config.Config, unsafeTrace bool) func() {
	if cfg.HasGovHostname() {
		return func() {}
	}

	return func() {
		u := ppanic.CollectPanic(cmd, nil, cfg)
		u.Report(ccloudv2.NewClient(cfg, unsafeTrace))
	}
}
