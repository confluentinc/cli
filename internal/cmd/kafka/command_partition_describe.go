package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

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
	partitionGetResp, resp, err := restClient.PartitionV3Api.GetKafkaPartition(restContext, clusterId, topic, partitionId)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	s := &struct {
		ClusterId   string
		TopicName   string
		PartitionId int32
		LeaderId    int32
	}{
		ClusterId:   partitionGetResp.ClusterId,
		TopicName:   partitionGetResp.TopicName,
		PartitionId: partitionGetResp.PartitionId,
		LeaderId:    parseLeaderId(partitionGetResp.Leader),
	}

	fields := []string{"ClusterId", "TopicName", "PartitionId", "LeaderId"}
	humanRenames := map[string]string{"ClusterId": "Cluster ID", "TopicName": "Topic Name", "PartitionId": "Partition ID", "LeaderId": "Leader ID"}
	structuredRenames := map[string]string{"ClusterId": "cluster_id", "TopicName": "topic_name", "PartitionId": "partition_id", "LeaderId": "leader_id"}

	return output.DescribeObject(cmd, s, fields, humanRenames, structuredRenames)
}
