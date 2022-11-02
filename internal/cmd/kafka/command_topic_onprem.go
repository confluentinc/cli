package kafka

// confluent kafka topic <commands>
import (
	"context"
	"fmt"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func getClusterIdForRestRequests(client *kafkarestv3.APIClient, ctx context.Context) (string, error) {
	clusters, resp, err := client.ClusterV3Api.ClustersGet(ctx)
	if err != nil {
		return "", kafkarest.NewError(client.GetConfig().BasePath, err, resp)
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
