package flink

import (
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
	if computePool == "" || err != nil {
		if c.Context.GetCurrentFlinkComputePool() == "" {
			return errors.NewErrorWithSuggestions("No compute pool set", "Please set a compute pool to be used. You can either set a default persitent compute pool \"confluent flink compute-pool use lfc-123\" or pass the flag \"--compute-pool lfcp-12345\".")
		} else {
			computePool = c.Context.GetCurrentFlinkComputePool()
		}
	}

	kafkaCluster, err := cmd.Flags().GetString("kafka-cluster")
	if kafkaCluster == "" || err != nil {
		if c.Context.KafkaClusterContext.GetActiveKafkaClusterId() != "" {
			kafkaCluster = c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
		}
	}

	enviromentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client.StartApp(enviromentId, resourceId, kafkaCluster, computePool, c.AuthToken,
		c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd),
		&types.ApplicationOptions{
			FLINK_GATEWAY_URL:        "https://flink.us-west-2.aws.devel.cpdev.cloud",
			HTTP_CLIENT_UNSAFE_TRACE: false,
			DEFAULT_PROPERTIES: map[string]string{
				"execution.runtime-mode": "streaming",
			},
			USER_AGENT: c.Version.UserAgent,
		})
	return nil
}
