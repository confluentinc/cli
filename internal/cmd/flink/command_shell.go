package flink

import (
	"net/url"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/load"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	client "github.com/confluentinc/cli/internal/pkg/flink/app"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/spf13/cobra"
)

func (c *command) newShellCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start Flink interactive SQL client.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.startFlinkSqlClient(prerunner, cmd, args)
		},
	}
	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")
	cmd.Flags().String("identity-pool", "", "Identity pool ID.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) authenticated(authenticated func(*cobra.Command, []string) error, cmd *cobra.Command) func() error {
	return func() error {
		cfg, err := load.LoadAndMigrate(v1.New())
		if err != nil {
			return err
		}
		auth := cfg.Context().State.AuthToken
		authRefreshToken := cfg.Context().State.AuthRefreshToken
		err = c.Context.UpdateAuthTokens(auth, authRefreshToken)
		if err != nil {
			return err
		}
		return authenticated(cmd, nil)
	}
}

func (c *command) startFlinkSqlClient(prerunner pcmd.PreRunner, cmd *cobra.Command, args []string) error {
	resourceId := c.Context.GetOrganization().GetResourceId()

	// Compute pool can be set as a flag or as default in the context
	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}
	if computePool == "" {
		if c.Context.GetCurrentFlinkComputePool() == "" {
			return errors.NewErrorWithSuggestions("No compute pool set", "Please set a compute pool to be used. You can either set a default persitent compute pool \"confluent flink compute-pool use lfc-123\" or pass the flag \"--compute-pool lfcp-12345\".")
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

	client.StartApp(flinkGatewayClient,
		c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd),
		types.ApplicationOptions{
			DefaultProperties: map[string]string{"execution.runtime-mode": "streaming"},
			FlinkGatewayUrl:   parsedUrl.String(),
			UnsafeTrace:       unsafeTrace,
			UserAgent:         c.Version.UserAgent,
			EnvId:             environmentId,
			OrgResourceId:     resourceId,
			KafkaClusterId:    cluster,
			ComputePoolId:     computePool,
			IdentityPoolId:    identityPool,
		})
	return nil
}
