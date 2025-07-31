package kafka

import (
	"fmt"
	"time"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/confluentinc/cli/v4/pkg/log"
)

// validate that a topic exists before attempting to produce/consume messages
func ValidateTopic(adminClient *ckgo.AdminClient, topic string) error {
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
		log.CliLogger.Tracef("validateTopic: Topic '%s' not visible in metadata, this could be due to ACL restrictions. Proceeding with operation to allow Kafka to handle authorization.", topic)
		// Instead of failing here, we log the situation and proceed. This allows:
		// 1. ALLOW topics to work properly and generate audit events
		// 2. DENY topics to fail at the consumer level and generate denial audit events
		// 3. Non-existent topics to fail with appropriate Kafka errors
		return nil
	}

	log.CliLogger.Tracef("validateTopic succeeded")
	return nil
}
