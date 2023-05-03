package flink

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	client "github.com/confluentinc/flink-sql-client"
	"github.com/confluentinc/flink-sql-client/pkg/types"

	"github.com/spf13/cobra"
)

func (c *command) newStartFlinkSqlClientCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start Flink interactive SQL client.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.startFlinkSqlClient(prerunner, cmd, args)
		},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Start Flink interactive SQL client.",
				Code: `confluent flink shell"`,
			},
		),
	}
	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")
	cmd.Flags().String("kafka-cluster", "", "Kafka cluster ID.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) authenticated(authenticated func(*cobra.Command, []string) error, cmd *cobra.Command) func() error {
	return func() error {
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

	kafkaClusterId, err := cmd.Flags().GetString("kafka-cluster")
	if kafkaClusterId == "" || err != nil {
		if c.Context.KafkaClusterContext.GetActiveKafkaClusterId() != "" {
			kafkaClusterId = c.Context.KafkaClusterContext.GetActiveKafkaClusterId()
		}
	}

	enviromentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	client.StartApp(enviromentId, resourceId, kafkaClusterId, computePool, c.AuthToken(),
		c.authenticated(prerunner.Authenticated(c.AuthenticatedCLICommand), cmd),
		&types.ApplicationOptions{
			FLINK_GATEWAY_URL:        "https://flink.us-west-2.aws.devel.cpdev.cloud",
			HTTP_CLIENT_UNSAFE_TRACE: false,
			DEFAULT_PROPERTIES: map[string]string{
				"execution.runtime-mode": "streaming",
			},
		})
	return nil
}
