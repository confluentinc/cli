package flink

import (
	"net/url"

	"github.com/spf13/cobra"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	client "github.com/confluentinc/cli/internal/pkg/flink/app"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

func (c *command) newShellCommand(prerunner pcmd.PreRunner) *cobra.Command {
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

	return cmd
}

func (c *command) authenticated(authenticated func(*cobra.Command, []string) error, cmd *cobra.Command, jwtValidator pcmd.JWTValidator) func() error {
	return func() error {
		cfg, err := load.LoadAndMigrate(v1.New())
		if err != nil {
			return err
		}

		authToken := cfg.Context().State.AuthToken
		authRefreshToken := cfg.Context().State.AuthRefreshToken
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
			flinkGatewayAuthToken, err := pauth.GetJwtTokenForV2Client(cfg.Context().GetState(), cfg.Context().GetPlatformServer())
			if err != nil {
				return err
			}
			flinkGatewayClient.AuthToken = flinkGatewayAuthToken
		}

		return nil
	}
}

func (c *command) startFlinkSqlClient(prerunner pcmd.PreRunner, cmd *cobra.Command) error {
	resourceId := c.Context.GetOrganization().GetResourceId()

	// Compute pool can be set as a flag or as default in the context
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

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	flinkComputePool, err := c.V2Client.DescribeFlinkComputePool(computePool, environmentId)
	if err != nil {
		return err
	}

	parsedUrl, err := url.Parse(flinkComputePool.Spec.GetHttpEndpoint())
	if err != nil {
		return err
	}
	parsedUrl.Path = ""

	unsafeTrace, err := c.Command.Flags().GetBool("unsafe-trace")
	if err != nil {
		return err
	}

	flinkGatewayClient, err := c.GetFlinkGatewayClient()
	if err != nil {
		return err
	}

	jwtValidator := pcmd.NewJWTValidator()

	client.StartApp(flinkGatewayClient,
		c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd, jwtValidator),
		types.ApplicationOptions{
			DefaultProperties: map[string]string{"execution.runtime-mode": "streaming"},
			FlinkGatewayUrl:   parsedUrl.String(),
			UnsafeTrace:       unsafeTrace,
			UserAgent:         c.Version.UserAgent,
			EnvironmentId:     environmentId,
			OrgResourceId:     resourceId,
			KafkaClusterId:    cluster,
			ComputePoolId:     computePool,
			IdentityPoolId:    identityPool,
		})
	return nil
}
