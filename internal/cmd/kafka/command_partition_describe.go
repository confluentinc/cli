package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
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
		Long:  "Describe a Kafka partition via Confluent Kafka REST.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe partition "1" for topic "my_topic".`,
				Code: "confluent kafka partition describe 1 --topic my_topic",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name to list partitions of.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("topic")

	return cmd
}

func (c *partitionCommand) describe(cmd *cobra.Command, args []string) error {
	partitionId, err := partitionIdFromArg(args)
	if err != nil {
		return err
	}

	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	partition, resp, err := restClient.PartitionV3Api.GetKafkaPartition(restContext, clusterId, topic, partitionId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	table := output.NewTable(cmd)
	table.Add(&partitionOut{
		ClusterId:   partition.ClusterId,
		TopicName:   partition.TopicName,
		PartitionId: partition.PartitionId,
		LeaderId:    parseLeaderId(partition.Leader),
	})
	return table.Print()
}
