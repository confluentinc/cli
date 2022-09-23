package kafka

// confluent kafka topic <commands>
import (
	"context"
	"fmt"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
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

func getClusterIdForRestRequests(client *kafkarestv3.APIClient, ctx context.Context) (string, error) {
	clusters, resp, err := client.ClusterV3Api.ClustersGet(ctx)
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
		log.CliLogger.Tracef("validateTopic: found topic %s", t.Topic)
		if topic == t.Topic {
			foundTopic = true // no break so that we see all topics from the above printout
		}
	}
	if !foundTopic {
		log.CliLogger.Tracef("validateTopic failed due to topic not being found in the client's topic list")
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicDoesNotExistOrMissingACLsErrorMsg, topic), fmt.Sprintf(errors.TopicDoesNotExistOrMissingACLsSuggestions, "<cluster-Id>", "<cluster-Id>", "<cluster-Id>"))
	}

	log.CliLogger.Tracef("validateTopic succeeded")
	return nil
}
