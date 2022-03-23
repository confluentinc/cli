package kafka

import (
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
	"net/http"
)

func (partitionCmd *partitionCommand) getReassignments(cmd *cobra.Command, args []string) error {
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
	var reassignmentListResp kafkarestv3.ReassignmentDataList
	var resp *http.Response
	if len(args) > 0 {
		partitionId, err := partitionIdFromArg(args)
		if err != nil {
			return err
		}
		if topic == "" {
			return errors.New(errors.SpecifyParitionIdWithTopicErrorMsg)
		}
		var reassignmentGetResp kafkarestv3.ReassignmentData
		reassignmentGetResp, resp, err = restClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReassignmentGet(restContext, clusterId, topic, partitionId)
		if err != nil {
			return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
		}
		if reassignmentGetResp.Kind != "" {
			reassignmentListResp.Data = []kafkarestv3.ReassignmentData{reassignmentGetResp}
		}
	} else if topic != "" {
		reassignmentListResp, resp, err = restClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsReassignmentGet(restContext, clusterId, topic)
	} else {
		reassignmentListResp, resp, err = restClient.PartitionApi.ClustersClusterIdTopicsPartitionsReassignmentGet(restContext, clusterId)
	}
	if err != nil {
		return kafkaRestError(restClient.GetConfig().BasePath, err, resp)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, []string{"ClusterId", "TopicName", "PartitionId", "AddingReplicas", "RemovingReplicas"}, []string{"Cluster ID", "Topic Name", "Partition ID", "Adding Replicas", "Removing Replicas"}, []string{"cluster_id", "topic_name", "partition_id", "adding_replicas", "removing_replicas"})
	if err != nil {
		return err
	}
	for _, data := range reassignmentListResp.Data {
		s := &struct {
			ClusterId        string
			TopicName        string
			PartitionId      int32
			AddingReplicas   []int32
			RemovingReplicas []int32
		}{
			ClusterId:        data.ClusterId,
			TopicName:        data.TopicName,
			PartitionId:      data.PartitionId,
			AddingReplicas:   data.AddingReplicas,
			RemovingReplicas: data.RemovingReplicas,
		}
		outputWriter.AddElement(s)
	}

	return outputWriter.Out()
}
