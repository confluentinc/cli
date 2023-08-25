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

func (c *command) newShellCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start Flink interactive SQL client.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.startFlinkSqlClient(prerunner, cmd)
		},
	}

	c.addComputePoolFlag(cmd)
	//TODO: remove as soon v1beta1 migration is complete (https://confluentinc.atlassian.net/browse/KFS-941)
	cmd.Flags().String("identity-pool", "", "Identity pool ID (deprecated).")
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("database", "", "The database which will be used as default database. When using Kafka, this is the cluster display name.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)
	if cfg.IsTest {
		cmd.Flags().Bool("fake-gateway", false, "Test the SQL client with fake gateway data.")
	}

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

		flinkGatewayClient, err := c.GetFlinkGatewayClient()
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
	// if the --fake-gateway flag is set, we start the client with a simulated gateway client that returns fake data
	fakeMode, _ := cmd.Flags().GetBool("fake-gateway")
	if fakeMode {
		client.StartApp(
			mock.NewFakeFlinkGatewayClient(),
			func() error { return nil },
			types.ApplicationOptions{
				Context:   c.Context,
				UserAgent: c.Version.UserAgent,
			}, func() {})
		return nil
	}

	resourceId := c.Context.GetOrganization().GetResourceId()

	environmentId, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environmentId == "" {
		if c.Context.GetCurrentEnvironment() == "" {
			return errors.NewErrorWithSuggestions("no environment provided", "Provide an environment with `confluent environment use env-123456` or `--environment`.")
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

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}
	if computePool == "" {
		if c.Context.GetCurrentFlinkComputePool() == "" {
			return errors.NewErrorWithSuggestions("no compute pool selected", "Select a compute pool with `confluent flink compute-pool use` or `--compute-pool`.")
		}
		computePool = c.Context.GetCurrentFlinkComputePool()
	}

	identityPool, err := cmd.Flags().GetString("identity-pool")
	if err != nil {
		return err
	}
	if identityPool == "" {
		identityPool = c.Context.GetCurrentIdentityPool()
	}

	serviceAccount, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}
	if serviceAccount == "" {
		serviceAccount = c.Context.GetCurrentServiceAccount()
	}

	if serviceAccount == "" && identityPool == "" {
		output.ErrPrintln("Warning: no service account provided. To ensure that your statements run continuously, " +
			"switch to using a service account instead of your user identity by providing a service account with `confluent iam service-account use` or `--service-account`. " +
			"Otherwise, statements will stop running after 4 hours.")
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

	flinkGatewayClient, err := c.GetFlinkGatewayClient()
	if err != nil {
		return err
	}

	jwtValidator := pcmd.NewJWTValidator()

	verbose, _ := cmd.Flags().GetCount("verbose")

	client.StartApp(flinkGatewayClient, c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd, jwtValidator), types.ApplicationOptions{
		Context:          c.Context,
		UnsafeTrace:      unsafeTrace,
		UserAgent:        c.Version.UserAgent,
		EnvironmentName:  catalog,
		EnvironmentId:    environmentId,
		OrgResourceId:    resourceId,
		Database:         database,
		ComputePoolId:    computePool,
		IdentityPoolId:   identityPool,
		ServiceAccountId: serviceAccount,
		Verbose:          verbose > 0,
	}, reportUsage(cmd, c.Config.Config, unsafeTrace))
	return nil
}

func reportUsage(cmd *cobra.Command, cfg *config.Config, unsafeTrace bool) func() {
	return func() {
		u := ppanic.CollectPanic(cmd, nil, cfg)
		u.Report(cfg.GetCloudClientV2(unsafeTrace))
	}
}
