package flink

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/auth"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	client "github.com/confluentinc/cli/v3/pkg/flink/app"
	"github.com/confluentinc/cli/v3/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
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
			dataplaneToken, err := auth.GetDataplaneToken(c.Context.GetState(), c.Context.GetPlatformServer())
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

	jwtValidator := pcmd.NewJWTValidator()

	verbose, _ := cmd.Flags().GetCount("verbose")

	cliVersion := c.Version.Version
	// due to a breaking change in the flink gateway, users using versions v3.36.0-v3.38.0 of the CLI
	// will not be able to submit SELECT statements anymore, so we should give them a warning
	if cliVersion == "v3.36.0" || cliVersion == "v3.37.0" || cliVersion == "v3.38.0" {
		output.ErrPrintf(c.Config.EnableColor, "[WARN] You're using an outdated version (%s) of the CLI."+
			" Please upgrade your CLI using `confluent update` or `brew upgrade confluentinc/tap/cli` to apply the latest changes.\n", cliVersion)
	}

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
	}

	client.StartApp(flinkGatewayClient, c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd, jwtValidator), opts, reportUsage(cmd, c.Config.Config, unsafeTrace))
	return nil
}

func reportUsage(cmd *cobra.Command, cfg *config.Config, unsafeTrace bool) func() {
	return func() {
		u := ppanic.CollectPanic(cmd, nil, cfg)
		u.Report(cfg.GetCloudClientV2(unsafeTrace))
	}
}
