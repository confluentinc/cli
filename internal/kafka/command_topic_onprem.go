package kafka

import (
	"fmt"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/log"
)

// validate that a topic exists before attempting to produce/consume messages
func ValidateTopic(adminClient *ckafka.AdminClient, topic string) error {
	timeout := 10 * time.Second
	metadata, err := adminClient.GetMetadata(nil, true, int(timeout.Milliseconds()))
	if err != nil {
		return fmt.Errorf("failed to obtain topics from client: %w", err)
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
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.TopicDoesNotExistOrMissingPermissionsErrorMsg, topic), fmt.Sprintf(errors.TopicDoesNotExistOrMissingPermissionsSuggestions, "<cluster-id>", "<cluster-id>", "<cluster-id>"))
	}

	log.CliLogger.Tracef("validateTopic succeeded")
	return nil
}
