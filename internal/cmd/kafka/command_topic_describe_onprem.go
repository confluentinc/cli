package kafka

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type PartitionData struct {
	TopicName              string  `human:"Topic" json:"topic" yaml:"topic"`
	PartitionId            int32   `human:"Partition" json:"partition" yaml:"partition"`
	LeaderBrokerId         int32   `human:"Leader" json:"leader" yaml:"leader"`
	ReplicaBrokerIds       []int32 `human:"Replicas" json:"replicas" yaml:"replicas"`
	InSyncReplicaBrokerIds []int32 `human:"ISR" json:"isr" yaml:"isr"`
}

type TopicData struct {
	TopicName         string            `json:"topic_name" yaml:"topic_name"`
	PartitionCount    int               `json:"partition_count" yaml:"partition_count"`
	ReplicationFactor int               `json:"replication_factor" yaml:"replication_factor"`
	Partitions        []*PartitionData  `json:"partitions" yaml:"partitions"`
	Configs           map[string]string `json:"config" yaml:"config"`
}

func (c *authenticatedTopicCommand) newDescribeCommandOnPrem() *cobra.Command {
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

func (c *authenticatedTopicCommand) describeOnPrem(cmd *cobra.Command, args []string) error {
	// Parse Args
	topicName := args[0]

	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}
	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	return DescribeTopicWithRESTClient(cmd, restClient, restContext, topicName, clusterId)
}

func DescribeTopicWithRESTClient(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, topicName, clusterId string) error {
	// Get partitions
	partitionsResp, resp, err := restClient.PartitionV3Api.ListKafkaPartitions(restContext, clusterId, topicName)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	} else if partitionsResp.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}
	topic := &TopicData{
		TopicName:      topicName,
		PartitionCount: len(partitionsResp.Data),
		Partitions:     make([]*PartitionData, len(partitionsResp.Data)),
	}
	for i, partitionResp := range partitionsResp.Data {
		// For each partition, get replicas
		replicasResp, resp, err := restClient.ReplicaApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(restContext, clusterId, topicName, partitionResp.PartitionId)
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		} else if replicasResp.Data == nil {
			return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
		}
		topic.Partitions[i] = &PartitionData{
			TopicName:              topicName,
			PartitionId:            partitionResp.PartitionId,
			ReplicaBrokerIds:       make([]int32, len(replicasResp.Data)),
			InSyncReplicaBrokerIds: make([]int32, 0, len(replicasResp.Data)),
		}
		for j, replicaResp := range replicasResp.Data {
			if replicaResp.IsLeader {
				topic.Partitions[i].LeaderBrokerId = replicaResp.BrokerId
			}
			topic.Partitions[i].ReplicaBrokerIds[j] = replicaResp.BrokerId
			if replicaResp.IsInSync {
				topic.Partitions[i].InSyncReplicaBrokerIds = append(topic.Partitions[i].InSyncReplicaBrokerIds, replicaResp.BrokerId)
			}
		}
		if i == 0 {
			topic.ReplicationFactor = len(replicasResp.Data)
		}
	}

	// Get configs
	configsResp, resp, err := restClient.ConfigsV3Api.ListKafkaTopicConfigs(restContext, clusterId, topicName)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	} else if configsResp.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}

	topic.Configs = make(map[string]string)
	for _, config := range configsResp.Data {
		if config.Value != nil {
			topic.Configs[config.Name] = *config.Value
		} else {
			topic.Configs[config.Name] = ""
		}
	}

	// Print topic info
	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, topic)
	}

	// Output partitions info
	output.Printf("Topic: %s\n", topic.TopicName)
	output.Printf("PartitionCount: %d\n", topic.PartitionCount)
	output.Printf("ReplicationFactor: %d\n\n", topic.ReplicationFactor)

	list := output.NewList(cmd)
	for _, partition := range topic.Partitions {
		list.Add(partition)
	}
	if err := list.Print(); err != nil {
		return err
	}
	output.Println()

	// Output config info
	output.Println("Configuration")
	output.Println()
	list = output.NewList(cmd)
	for name, value := range topic.Configs {
		list.Add(&configOut{
			Name:  name,
			Value: value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
