package flink

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	client "github.com/confluentinc/cli/internal/pkg/flink/app"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

func (c *command) newShellCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start Flink interactive SQL client.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.startFlinkSqlClient(prerunner, cmd)
		},
	}

	c.addComputePoolFlag(cmd)
	cmd.Flags().String("identity-pool", "", "Identity pool ID.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
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

		jwtCtx := &v1.Context{State: &v1.ContextState{AuthToken: flinkGatewayClient.AuthToken}}
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
			})
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

	environment, err := c.V2Client.GetOrgEnvironment(environmentId)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), "List available environments with `confluent environment list`.")
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
		if c.Context.GetCurrentIdentityPool() == "" {
			return errors.NewErrorWithSuggestions("no identity pool set", "Set a persistent identity pool with `confluent iam pool use` or pass the `--identity-pool` flag.")
		}
		identityPool = c.Context.GetCurrentIdentityPool()
	}

	cluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return err
	}
	if cluster == "" {
		if c.Context.KafkaClusterContext.GetActiveKafkaClusterId() != "" {
			cluster = c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
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

	// use name of environment and not id
	client.StartApp(
		flinkGatewayClient,
		c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd, jwtValidator),
		types.ApplicationOptions{
			Context:         c.Context,
			UnsafeTrace:     unsafeTrace,
			UserAgent:       c.Version.UserAgent,
			EnvironmentName: environment.GetDisplayName(),
			EnvironmentId:   environmentId,
			OrgResourceId:   resourceId,
			KafkaClusterId:  cluster,
			ComputePoolId:   computePool,
			IdentityPoolId:  identityPool,
			Verbose:         verbose > 0,
		})
	return nil
}
