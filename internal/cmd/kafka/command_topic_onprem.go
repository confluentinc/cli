package kafka

// confluent kafka topic <commands>
import (
	"context"
	"fmt"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type PartitionData struct {
	TopicName              string  `json:"topic" yaml:"topic"`
	PartitionId            int32   `json:"partition" yaml:"partition"`
	LeaderBrokerId         int32   `json:"leader" yaml:"leader"`
	ReplicaBrokerIds       []int32 `json:"replicas" yaml:"replicas"`
	InSyncReplicaBrokerIds []int32 `json:"isr" yaml:"isr"`
}

type TopicData struct {
	TopicName         string            `json:"topic_name" yaml:"topic_name"`
	PartitionCount    int               `json:"partition_count" yaml:"partition_count"`
	ReplicationFactor int               `json:"replication_factor" yaml:"replication_factor"`
	Partitions        []PartitionData   `json:"partitions" yaml:"partitions"`
	Configs           map[string]string `json:"config" yaml:"config"`
}

// Register each of the verbs and expected args
func (c *authenticatedTopicCommand) onPremInit() {
	c.AddCommand(c.newListCommandOnPrem())
	c.AddCommand(c.newCreateCommandOnPrem())
	c.AddCommand(c.newDeleteCommandOnPrem())
	c.AddCommand(c.newUpdateCommandOnPrem())
	c.AddCommand(c.newDescribeCommandOnPrem())
	c.AddCommand(c.newProduceCommandOnPrem())
	c.AddCommand(c.newConsumeCommandOnPrem())
}

func getClusterIdForRestRequests(client *kafkarestv3.APIClient, ctx context.Context) (string, error) {
	clusters, resp, err := client.ClusterApi.ClustersGet(ctx)
	if err != nil {
		return "", kafkaRestError(client.GetConfig().BasePath, err, resp)
	}
	if clusters.Data == nil || len(clusters.Data) == 0 {
		return "", errors.NewErrorWithSuggestions(errors.NoClustersFoundErrorMsg, errors.NoClustersFoundSuggestions)
	}
	clusterId := clusters.Data[0].ClusterId
	return clusterId, nil
}

// validate that a topic exists before attempting to produce/consume messages
func (c *authenticatedTopicCommand) validateTopic(adminClient *ckafka.AdminClient, topic string) error {
	timeout := 10 * time.Second
	metadata, err := adminClient.GetMetadata(nil, true, int(timeout.Milliseconds()))
	if err != nil {
		return fmt.Errorf("failed to obtain topics from client: %v", err)
	}

	var foundTopic bool
	for _, t := range metadata.Topics {
		c.logger.Tracef("validateTopic: found topic " + t.Topic)
		if topic == t.Topic {
			foundTopic = true // no break so that we see all topics from the above printout
		}
	}
	if !foundTopic {
		c.logger.Tracef("validateTopic failed due to topic not being found in the client's topic list")
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicDoesNotExistOrMissingACLsErrorMsg, topic), fmt.Sprintf(errors.TopicDoesNotExistOrMissingACLsSuggestions, "<cluster-Id>", "<cluster-Id>", "<cluster-Id>"))
	}

	c.logger.Tracef("validateTopic succeeded")
	return nil
}
