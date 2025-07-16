package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type partitionOut struct {
	Cluster   string `human:"Cluster" serialized:"cluster"`
	TopicName string `human:"Topic Name" serialized:"topic_name"`
	Id        int32  `human:"ID" serialized:"id"`
	Leader    int32  `human:"Leader" serialized:"leader"`
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
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
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

	kafkaREST, err := c.GetKafkaREST(cmd)
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
		Cluster:   partition.GetClusterId(),
		TopicName: partition.GetTopicName(),
		Id:        partition.GetPartitionId(),
		Leader:    parseLeaderId(partition.GetLeader().Related),
	})
	return table.Print()
}
