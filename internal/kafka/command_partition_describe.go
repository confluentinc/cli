package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type partitionOut struct {
	ClusterId   string `human:"Cluster" serialized:"cluster_id"`
	TopicName   string `human:"Topic Name" serialized:"topic_name"`
	PartitionId int32  `human:"Partition ID" serialized:"partition_id"`
	LeaderId    int32  `human:"Leader ID" serialized:"leader_id"`
}

func (c *partitionCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Kafka partition.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe partition "1" for topic "my_topic".`,
				Code: "confluent kafka partition describe 1 --topic my_topic",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name to describe a partition of.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("topic"))

	return cmd
}

func (c *partitionCommand) describe(cmd *cobra.Command, args []string) error {
	partitionId, err := partitionIdFromArg(args)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	partition, err := kafkaREST.CloudClient.GetKafkaPartition(topic, partitionId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&partitionOut{
		ClusterId:   partition.GetClusterId(),
		TopicName:   partition.GetTopicName(),
		PartitionId: partition.GetPartitionId(),
		LeaderId:    parseLeaderId(partition.GetLeader().Related),
	})
	return table.Print()
}
