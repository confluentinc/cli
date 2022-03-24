package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (partitionCmd *partitionCommand) newDescribeCommand() *cobra.Command {
	describeCmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Kafka partition.",
		Long:  "Describe a Kafka partition via Confluent Kafka REST.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(partitionCmd.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe partition "1" for "my_topic".`,
				Code: "confluent kafka partition describe 1 --topic my_topic",
			}),
	}
	describeCmd.Flags().String("topic", "", "Topic name to list partitions of.")
	describeCmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(describeCmd)
	_ = describeCmd.MarkFlagRequired("topic")
	partitionCmd.AddCommand(describeCmd)
	return describeCmd
}

func (partitionCmd *partitionCommand) describe(cmd *cobra.Command, args []string) error {
	partitionId, err := partitionIdFromArg(args)
	if err != nil {
		return err
	}
	restClient, restContext, err := initKafkaRest(partitionCmd.AuthenticatedCLICommand, cmd)
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
	return output.DescribeObject(cmd, s, []string{"ClusterId", "TopicName", "PartitionId", "LeaderId"}, map[string]string{"ClusterId": "Cluster ID", "TopicName": "Topic Name", "PartitionId": "Partition ID", "LeaderId": "Leader ID"}, map[string]string{"ClusterId": "cluster_id", "TopicName": "topic_name", "PartitionId": "partition_id", "LeaderId": "leader_id"})
}
