package flink

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/auth"
	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	client "github.com/confluentinc/cli/v3/pkg/flink/app"
	"github.com/confluentinc/cli/v3/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	ppanic "github.com/confluentinc/cli/v3/pkg/panic-recovery"
)

// If we set this const useFakeGateway to true, we start the client with a simulated gateway client that returns fake data. This is used for debugging.
const useFakeGateway = false

func (c *command) newShellCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start Flink interactive SQL client.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.startFlinkSqlClient(prerunner, cmd)
		},
	}

	c.addComputePoolFlag(cmd)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	c.addDatabaseFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	cmd.Flags().Bool("enable-lsp", false, "Enables the flink language service integration (experimental).")

	return cmd
}

func (c *command) authenticated(authenticated func(*cobra.Command, []string) error, cmd *cobra.Command, jwtValidator pcmd.JWTValidator) func() error {
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
	if useFakeGateway {
		client.StartApp(
			mock.NewFakeFlinkGatewayClient(),
			func() error { return nil },
			types.ApplicationOptions{
				Context:   c.Context,
				UserAgent: c.Version.UserAgent,
			}, func() {})
		return nil
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
	if serviceAccount == "" {
		output.ErrPrintln(c.Config.EnableColor, serviceAccountWarning)
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

	lspEnabled, err := cmd.Flags().GetBool("enable-lsp")
	if err != nil {
		return err
	}

	var lspBaseUrl string
	if lspEnabled {
		lspBaseUrl, err = c.getFlinkLanguageServiceUrl(flinkGatewayClient)
		if err != nil {
			log.CliLogger.Warnf("Shell won't connect to language service. Error getting language service url: %v\n", err)
			return err
		}
	}

	jwtValidator := pcmd.NewJWTValidator()

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
		LSPEnabled:       lspEnabled,
		LSPBaseUrl:       lspBaseUrl,
	}

	client.StartApp(flinkGatewayClient, c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd, jwtValidator), opts, reportUsage(cmd, c.Config, unsafeTrace))
	return nil
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
	return func() {
		u := ppanic.CollectPanic(cmd, nil, cfg)
		u.Report(ccloudv2.NewClient(cfg, unsafeTrace))
	}
}
