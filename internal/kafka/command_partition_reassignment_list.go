package kafka

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type getReassignmentOutHuman struct {
	ClusterId        string `human:"Cluster"`
	TopicName        string `human:"Topic Name"`
	PartitionId      int32  `human:"Partition ID"`
	AddingReplicas   string `human:"Adding Replicas"`
	RemovingReplicas string `human:"Removing Replicas"`
}

type getReassignmentOutSerialized struct {
	ClusterId        string  `serialized:"cluster_id"`
	TopicName        string  `serialized:"topic_name"`
	PartitionId      int32   `serialized:"partition_id"`
	AddingReplicas   []int32 `serialized:"adding_replicas"`
	RemovingReplicas []int32 `serialized:"removing_replicas"`
}

func (c *partitionCommand) newReassignmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [id]",
		Short: "List ongoing partition reassignments.",
		Long:  "List ongoing partition reassignments for a given cluster, topic, or partition.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.reassignmentList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all partition reassignments for the Kafka cluster.",
				Code: "confluent kafka partition reassignment list",
			},
			examples.Example{
				Text: `List partition reassignments for topic "my_topic".`,
				Code: "confluent kafka partition reassignment list --topic my_topic",
			},
			examples.Example{
				Text: `List partition reassignments for partition "1" of topic "my_topic".`,
				Code: "confluent kafka partition reassignment list 1 --topic my_topic",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name to search by.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *partitionCommand) reassignmentList(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}
	var reassignments kafkarestv3.ReassignmentDataList
	var resp *http.Response
	if len(args) > 0 {
		partitionId, err := partitionIdFromArg(args)
		if err != nil {
			return err
		}
		if topic == "" {
			return fmt.Errorf("must specify topic along with partition ID")
		}
		var reassignmentGetResp kafkarestv3.ReassignmentData
		reassignmentGetResp, resp, err = restClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReassignmentGet(restContext, clusterId, topic, partitionId)
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		if reassignmentGetResp.Kind != "" {
			reassignments.Data = []kafkarestv3.ReassignmentData{reassignmentGetResp}
		}
	} else if topic != "" {
		reassignments, resp, err = restClient.PartitionApi.ClustersClusterIdTopicsTopicNamePartitionsReassignmentGet(restContext, clusterId, topic)
	} else {
		reassignments, resp, err = restClient.PartitionApi.ClustersClusterIdTopicsPartitionsReassignmentGet(restContext, clusterId)
	}
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, reassignment := range reassignments.Data {
		if output.GetFormat(cmd) == output.Human {
			list.Add(&getReassignmentOutHuman{
				ClusterId:        reassignment.ClusterId,
				TopicName:        reassignment.TopicName,
				PartitionId:      reassignment.PartitionId,
				AddingReplicas:   join(reassignment.AddingReplicas),
				RemovingReplicas: join(reassignment.RemovingReplicas),
			})
		} else {
			list.Add(&getReassignmentOutSerialized{
				ClusterId:        reassignment.ClusterId,
				TopicName:        reassignment.TopicName,
				PartitionId:      reassignment.PartitionId,
				AddingReplicas:   reassignment.AddingReplicas,
				RemovingReplicas: reassignment.RemovingReplicas,
			})
		}
	}
	return list.Print()
}

func join(replicas []int32) string {
	s := make([]string, len(replicas))
	for i, replica := range replicas {
		s[i] = strconv.Itoa(int(replica))
	}
	return strings.Join(s, ", ")
}
