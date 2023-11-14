package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *partitionCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka partitions.",
		Long:  "List the partitions that belong to a specified topic.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the partitions of topic "my_topic".`,
				Code: "confluent kafka partition list --topic my_topic",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name to list partitions of.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("topic"))

	return cmd
}

func (c *partitionCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	partitions, err := kafkaREST.CloudClient.ListKafkaPartitions(topic)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, partition := range partitions {
		list.Add(&partitionOut{
			ClusterId:   partition.GetClusterId(),
			TopicName:   partition.GetTopicName(),
			PartitionId: partition.GetPartitionId(),
			LeaderId:    parseLeaderId(partition.Leader.GetRelated()),
		})
	}
	return list.Print()
}
