package kafka

import (
	"fmt"
	"io"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/log"

	"github.com/Shopify/sarama"
)

// This is a nasty side effect of Sarama using a global logger
func InitSarama(logger *log.Logger) {
	sarama.Logger = newLogAdapter(logger)
}

// NewSaramaConsumer returns a sarama.ConsumerGroup configured for the CLI config
func NewSaramaConsumer(group string, kafka *config.KafkaClusterConfig, clientID string, beginning bool) (sarama.ConsumerGroup, error) {
	client, err := sarama.NewClient(strings.Split(kafka.Bootstrap, ","), saramaConf(kafka, clientID, beginning))
	if err != nil {
		return nil, err
	}
	return sarama.NewConsumerGroupFromClient(group, client)
}

// NewSaramaProducer returns a sarama.ClusterProducer configured for the CLI config
func NewSaramaProducer(kafka *config.KafkaClusterConfig, clientID string) (sarama.SyncProducer, error) {
	return sarama.NewSyncProducer(strings.Split(kafka.Bootstrap, ","), saramaConf(kafka, clientID, false))
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
func saramaConf(kafka *config.KafkaClusterConfig, clientID string, beginning bool) *sarama.Config {
	saramaConf := sarama.NewConfig()
	saramaConf.ClientID = clientID
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

// Just logs all Sarama logs at the debug level
// We don't use hclog.StandardLogger() because that prints at INFO level
type logAdapter struct {
	logger *log.Logger
}

func newLogAdapter(logger *log.Logger) *logAdapter {
	return &logAdapter{logger: logger}
}

func (l *logAdapter) Print(a ...interface{}) {
	l.log(fmt.Sprint(a...))
}

func (l *logAdapter) Println(a ...interface{}) {
	l.log(fmt.Sprint(a...))
}

func (l *logAdapter) Printf(format string, a ...interface{}) {
	l.log(fmt.Sprintf(format, a...))
}

func (l *logAdapter) log(msg string) {
	// This is how hclog.StandardLogger works as well; it fixes the unnecessary extra newlines
	msg = string(strings.TrimRight(msg, " \t\n"))
	l.logger.Log("msg", msg)
}
