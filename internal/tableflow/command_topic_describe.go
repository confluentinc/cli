package tableflow

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
)

func (c *command) newTopicDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <name>",
		Short:             "Describe a topic.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validTopicArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe a Tableflow topic "my-tableflow-topic" related to a Kafka cluster "lkc-123456".`,
				Code: "confluent tableflow topic describe my-tableflow-topic --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	name := args[0]

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	topic, err := c.V2Client.GetTableflowTopic(environmentId, cluster.GetId(), name)
	if err != nil {
		return err
	}

	return printTopicTable(cmd, topic)
}
