package kafka

import (
	"fmt"
	"io"
	"strings"

	"github.com/Shopify/sarama"

	"github.com/confluentinc/cli/internal/pkg/config"
)

// NewSaramaConsumer returns a sarama.ConsumerGroup configured for the CLI config
func NewSaramaConsumer(group string, kafka *config.KafkaClusterConfig, beginning bool) (sarama.ConsumerGroup, error) {
	client, err := sarama.NewClient(strings.Split(kafka.Bootstrap, ","), saramaConf(kafka, beginning))
	if err != nil {
		return nil, err
	}
	return sarama.NewConsumerGroupFromClient(group, client)
}

// NewSaramaProducer returns a sarama.ClusterProducer configured for the CLI config
func NewSaramaProducer(kafka *config.KafkaClusterConfig) (sarama.SyncProducer, error) {
	return sarama.NewSyncProducer(strings.Split(kafka.Bootstrap, ","), saramaConf(kafka, false))
}

// GroupHandler instances are used to handle individual topic-partition claims.
type GroupHandler struct {
	Out io.Writer
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (*GroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (*GroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (h *GroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		_, err := fmt.Fprintln(h.Out, string(msg.Value))
		if err != nil {
			return err
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

// saramaConf converts KafkaClusterConfig to sarama.Config
func saramaConf(kafka *config.KafkaClusterConfig, beginning bool) *sarama.Config {
	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V1_1_0_0
	saramaConf.Net.TLS.Enable = true
	saramaConf.Net.SASL.Enable = true
	saramaConf.Net.SASL.User = kafka.APIKey
	saramaConf.Net.SASL.Password = kafka.APIKeys[kafka.APIKey].Secret

	saramaConf.Producer.Return.Successes = true
	saramaConf.Producer.Return.Errors = true

	if beginning {
		saramaConf.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		saramaConf.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	return saramaConf
}
