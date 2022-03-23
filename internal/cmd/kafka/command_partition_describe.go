package kafka

import (
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

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
