package flink

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	client "github.com/confluentinc/flink-sql-client"
	application "github.com/confluentinc/flink-sql-client/pkg/controller"
	"github.com/spf13/cobra"
)

func (c *command) newStartFlinkSqlClientCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Start Flink interactive SQL client.",
		RunE:  c.startFlinkSqlClient,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Start Flink interactive SQL client.",
				Code: `confluent flink shell"`,
			},
		),
	}
	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) startFlinkSqlClient(cmd *cobra.Command, args []string) error {
	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}
	resourceId := c.Context.GetOrganization().GetResourceId()
	ctx := c.Context.Config.Context()
	kafkaClusterId := ctx.KafkaClusterContext.GetActiveKafkaClusterId()

	client.StartApp(c.EnvironmentId(), resourceId, kafkaClusterId, computePool, c.AuthToken(), &application.ApplicationOptions{MOCK_STATEMENTS_OUTPUT_DEMO: true})
	return nil
}
