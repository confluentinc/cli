package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type ReplicaOut struct {
	ClusterId   string `human:"Cluster" serialized:"cluster_id"`
	TopicName   string `human:"Topic Name" serialized:"topic_name"`
	PartitionId int32  `human:"Partition ID" serialized:"partition_id"`
	BrokerId    int32  `human:"Broker ID" serialized:"broker_id"`
	IsLeader    bool   `human:"Leader" serialized:"is_leader"`
	IsInIsr     bool   `human:"In ISR" serialized:"is_in_isr"`
}

func (c *replicaCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <topic>",
		Short: "List Kafka replicas.",
		Long:  "List partition-replicas for a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List the replicas for partition 1 of topic "my-topic".`,
				Code: "confluent kafka replica list my-topic --partition 1",
			},
			examples.Example{
				Text: `List the replicas for topic "my-topic".`,
				Code: "confluent kafka replica list my-topic",
			},
		),
	}

	cmd.Flags().Int32("partition", -1, "Partition ID.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *replicaCommand) list(cmd *cobra.Command, args []string) error {
	topic := args[0]

	partitionId, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	var replicas []kafkarestv3.ReplicaData
	if partitionId != -1 {
		partitionReplicas, resp, err := restClient.ReplicaApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(restContext, clusterId, topic, partitionId)
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		} else if partitionReplicas.Data == nil {
			return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
		}
		replicas = partitionReplicas.Data
	} else {
		partitions, resp, err := restClient.PartitionV3Api.ListKafkaPartitions(restContext, clusterId, topic)
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		} else if partitions.Data == nil {
			return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
		}
		for _, partition := range partitions.Data {
			partitionReplicas, resp, err := restClient.ReplicaApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(restContext, clusterId, topic, partition.PartitionId)
			if err != nil {
				return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
			} else if partitionReplicas.Data == nil {
				return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
			}
			replicas = append(replicas, partitionReplicas.Data...)
		}
	}

	list := output.NewList(cmd)
	for _, replica := range replicas {
		list.Add(&ReplicaOut{
			ClusterId:   replica.ClusterId,
			TopicName:   replica.TopicName,
			PartitionId: replica.PartitionId,
			BrokerId:    replica.BrokerId,
			IsLeader:    replica.IsLeader,
			IsInIsr:     replica.IsInSync,
		})
	}
	return list.Print()
}
