package kafka

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
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
	// Parse Args
	topicName := args[0]

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return DescribeTopic(cmd, restClient, restContext, topicName, clusterId)
}

func DescribeTopic(cmd *cobra.Command, restClient *kafkarestv3.APIClient, restContext context.Context, topicName, clusterId string) error {
	// Get partitions
	partitions, resp, err := restClient.PartitionV3Api.ListKafkaPartitions(restContext, clusterId, topicName)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	} else if partitions.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}

	topic := &TopicData{
		TopicName:      topicName,
		PartitionCount: len(partitions.Data),
		Partitions:     make([]*PartitionData, len(partitions.Data)),
	}

	for i, partition := range partitions.Data {
		// For each partition, get replicas
		replicas, resp, err := restClient.ReplicaApi.ClustersClusterIdTopicsTopicNamePartitionsPartitionIdReplicasGet(restContext, clusterId, topicName, partition.PartitionId)
		if err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		} else if replicas.Data == nil {
			return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
		}
		topic.Partitions[i] = &PartitionData{
			TopicName:              topicName,
			PartitionId:            partition.PartitionId,
			ReplicaBrokerIds:       make([]int32, len(replicas.Data)),
			InSyncReplicaBrokerIds: make([]int32, 0, len(replicas.Data)),
		}
		for j, replica := range replicas.Data {
			if replica.IsLeader {
				topic.Partitions[i].LeaderBrokerId = replica.BrokerId
			}
			topic.Partitions[i].ReplicaBrokerIds[j] = replica.BrokerId
			if replica.IsInSync {
				topic.Partitions[i].InSyncReplicaBrokerIds = append(topic.Partitions[i].InSyncReplicaBrokerIds, replica.BrokerId)
			}
		}
		if i == 0 {
			topic.ReplicationFactor = len(replicas.Data)
		}
	}

	// Get configs
	configs, resp, err := restClient.ConfigsV3Api.ListKafkaTopicConfigs(restContext, clusterId, topicName)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	} else if configs.Data == nil {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, errors.InternalServerErrorSuggestions)
	}

	topic.Configs = make(map[string]string)
	for _, config := range configs.Data {
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
	output.Printf(false, "Topic: %s\n", topic.TopicName)
	output.Printf(false, "Partition Count: %d\n", topic.PartitionCount)
	output.Printf(false, "Replication Factor: %d\n\n", topic.ReplicationFactor)

	list := output.NewList(cmd)
	for _, partition := range topic.Partitions {
		list.Add(partition)
	}
	if err := list.Print(); err != nil {
		return err
	}
	output.Println(false, "")

	// Output config info
	output.Println(false, "Configuration")
	output.Println(false, "")
	list = output.NewList(cmd)
	for name, value := range topic.Configs {
		list.Add(&broker.ConfigOut{
			Name:  name,
			Value: value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
