package kafka

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type PartitionData struct {
	Partition int32   `human:"Partition" serialized:"partition"`
	Leader    int32   `human:"Leader" serialized:"leader"`
	Replicas  []int32 `human:"Replicas" serialized:"replicas"`
	Isr       []int32 `human:"ISR" serialized:"isr"`
}

type describeOutOnPrem struct {
	TopicOut
	Partitions []*PartitionData `serialized:"partitions"`
}

func (c *command) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describeOnPrem,
		Short: "Describe a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "my_topic" topic for the specified cluster (providing embedded Kafka REST Proxy endpoint).`,
				Code: "confluent kafka topic describe my_topic --url http://localhost:8090/kafka",
			},

			examples.Example{
				Text: `Describe the "my_topic" topic for the specified cluster (providing Kafka REST Proxy endpoint).`,
				Code: "confluent kafka topic describe my_topic --url http://localhost:8082",
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describeOnPrem(cmd *cobra.Command, args []string) error {
	topicName := args[0]

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return DescribeTopic(cmd, restClient, restContext, topicName, clusterId)
}

func DescribeTopic(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, topicName, clusterId string) error {
	topic, resp, err := restClient.TopicV3Api.GetKafkaTopic(restContext, clusterId, topicName)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	topicDescribe := &TopicOut{
		Name:              topic.TopicName,
		IsInternal:        topic.IsInternal,
		ReplicationFactor: topic.ReplicationFactor,
		PartitionCount:    topic.PartitionsCount,
	}

	partitions, resp, err := restClient.PartitionV3Api.ListKafkaPartitions(restContext, clusterId, topicName)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	} else if partitions.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}

	partitionList := make([]*PartitionData, len(partitions.Data))

	for i, partition := range partitions.Data {
		replicas, resp, err := restClient.ReplicaApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(restContext, clusterId, topicName, partition.PartitionId)
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		} else if replicas.Data == nil {
			return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
		}
		partitionList[i] = &PartitionData{
			Partition: partition.PartitionId,
			Replicas:  make([]int32, len(replicas.Data)),
			Isr:       make([]int32, 0, len(replicas.Data)),
		}
		for j, replica := range replicas.Data {
			if replica.IsLeader {
				partitionList[i].Leader = replica.BrokerId
			}
			partitionList[i].Replicas[j] = replica.BrokerId
			if replica.IsInSync {
				partitionList[i].Isr = append(partitionList[i].Isr, replica.BrokerId)
			}
		}
	}

	if output.GetFormat(cmd).IsSerialized() {
		table := output.NewTable(cmd)
		table.Add(&describeOutOnPrem{
			TopicOut:   *topicDescribe,
			Partitions: partitionList,
		})
		return table.Print()
	}

	table := output.NewTable(cmd)
	table.Add(topicDescribe)
	if err := table.Print(); err != nil {
		return err
	}

	output.Println(false, "")

	list := output.NewList(cmd)
	for _, partition := range partitionList {
		list.Add(partition)
	}
	return list.Print()
}
