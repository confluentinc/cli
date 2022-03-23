package kafka

import (
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (partitionCmd *partitionCommand) list(cmd *cobra.Command, _ []string) error {
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
	partitionListResp, resp, err := restClient.PartitionV3Api.ListKafkaPartitions(restContext, clusterId, topic)
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}
	partitionDatas := partitionListResp.Data

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"ClusterId", "TopicName", "PartitionId", "LeaderId"}, []string{"Cluster ID", "Topic Name", "Partition ID", "Leader ID"}, []string{"cluster_id", "topic_name", "partition_id", "leader_id"})
	if err != nil {
		return err
	}
	for _, partition := range partitionDatas {
		s := &struct {
			ClusterId   string
			TopicName   string
			PartitionId int32
			LeaderId    int32
		}{
			ClusterId:   partition.ClusterId,
			TopicName:   partition.TopicName,
			PartitionId: partition.PartitionId,
			LeaderId:    parseLeaderId(partition.Leader),
		}
		outputWriter.AddElement(s)
	}

	return outputWriter.Out()
}
