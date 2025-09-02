package unifiedstreammanager

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newKafkaDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <usm-cluster-id>",
		Short:             "Describe a Kafka cluster.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validKafkaArgs),
		RunE:              c.describeKafka,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe a Confluent Platform Kafka cluster with the ID 4k0R9d1GTS5tI9f4Y2xZ0Q.",
				Code: "confluent unified-stream-manager kafka describe 4k0R9d1GTS5tI9f4Y2xZ0Q",
			},
		),
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describeKafka(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := c.V2Client.GetUsmKafkaCluster(args[0], environmentId)
	if err != nil {
		return err
	}

	return printKafkaTable(cmd, cluster)
}
