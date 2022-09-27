package kafka

import (
	"net/http"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type reassignmentsOut struct {
	ClusterId        string  `human:"Cluster ID" serialized:"cluster_id"`
	TopicName        string  `human:"Topic Name" serialized:"topic_name"`
	PartitionId      int32   `human:"Partition ID" serialized:"partition_id"`
	AddingReplicas   []int32 `human:"Adding Replicas" serialized:"adding_replicas"`
	RemovingReplicas []int32 `human:"Removing Replicas" serialized:"removing_replicas"`
}

func (c *partitionCommand) newGetReassignmentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-reassignments [id]",
		Short: "Get ongoing replica reassignments.",
		Long:  "Get ongoing replica reassignments for a given cluster, topic, or partition via Confluent Kafka REST.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.getReassignments,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get all replica reassignments for the Kafka cluster.",
				Code: "confluent kafka partition get-reassignments",
			},
			examples.Example{
				Text: `Get replica reassignments for topic "my_topic".`,
				Code: "confluent kafka partition get-reassignments --topic my_topic",
			},
			examples.Example{
				Text: `Get replica reassignments for partition "1" of topic "my_topic".`,
				Code: "confluent kafka partition get-reassignments 1 --topic my_topic",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name to search by.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *partitionCommand) getReassignments(cmd *cobra.Command, args []string) error {
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
	var reassignmentListResp kafkarestv3.ReassignmentDataList
	var resp *http.Response
	if len(args) > 0 {
		partitionId, err := partitionIdFromArg(args)
		if err != nil {
			return err
		}
		if topic == "" {
			return errors.New(errors.SpecifyPartitionIdWithTopicErrorMsg)
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

	list := output.NewList(cmd)
	for _, reassignment := range reassignmentListResp.Data {
		list.Add(&reassignmentsOut{
			ClusterId:        reassignment.ClusterId,
			TopicName:        reassignment.TopicName,
			PartitionId:      reassignment.PartitionId,
			AddingReplicas:   reassignment.AddingReplicas,
			RemovingReplicas: reassignment.RemovingReplicas,
		})
	}
	return list.Print()
}
