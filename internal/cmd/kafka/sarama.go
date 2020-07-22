package kafka

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
	serdes "github.com/confluentinc/cli/internal/pkg/serdes"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/Shopify/sarama"
)

// This is a nasty side effect of Sarama using a global logger
func InitSarama(logger *log.Logger) {
	sarama.Logger = newLogAdapter(logger)
}

// NewSaramaConsumer returns a sarama.ConsumerGroup configured for the CLI config
func NewSaramaConsumer(group string, kafka *v1.KafkaClusterConfig, clientID string, beginning bool) (sarama.ConsumerGroup, error) {
	conf, err := saramaConf(kafka, clientID, beginning)
	if err != nil {
		return nil, err
	}
	client, err := sarama.NewClient(strings.Split(kafka.Bootstrap, ","), conf)
	if err != nil {
		return nil, err
	}
	return sarama.NewConsumerGroupFromClient(group, client)
}

// NewSaramaProducer returns a sarama.ClusterProducer configured for the CLI config
func NewSaramaProducer(kafka *v1.KafkaClusterConfig, clientID string) (sarama.SyncProducer, error) {
	conf, err := saramaConf(kafka, clientID, false)
	if err != nil {
		return nil, err
	}
	return sarama.NewSyncProducer(strings.Split(kafka.Bootstrap, ","), conf)
}

type ConsumerProperties struct {
	PrintKey   bool
	Delimiter  string
	SchemaPath string
}

// GroupHandler instances are used to handle individual topic-partition claims.
type GroupHandler struct {
	SrClient   *srsdk.APIClient
	Ctx        context.Context
	Format     string
	Out        io.Writer
	Properties ConsumerProperties
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
		value := msg.Value
		// For messages with schema, first byte is magic byte 0x0.
		_ = value[0]
		if h.Properties.PrintKey {
			key := msg.Key
			var keyString string
			if len(key) == 0 {
				keyString = "null"
			} else {
				keyString = string(key)
			}
			_, err := fmt.Fprint(h.Out, keyString+h.Properties.Delimiter)
			if err != nil {
				return err
			}
		}

		deserializationProvider, err := serdes.GetDeserializationProvider(h.Format)
		if err != nil {
			return err
		}

		if h.Format != "RAW" {
			schemaPath, err := h.RequestSchema(value)
			if err != nil {
				return err
			}
			// Message body is encoded after 5 bytes of meta information.
			value = value[5:]
			err = deserializationProvider.LoadSchema(schemaPath)
			if err != nil {
				return err
			}
		}
		jsonMessage, err := serdes.Deserialize(deserializationProvider, value)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintln(h.Out, jsonMessage)
		if err != nil {
			return err
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (h *GroupHandler) RequestSchema(value []byte) (string, error) {
	// Retrieve schema from cluster only if schema is specified.
	schemaID := int32(binary.BigEndian.Uint32(value[1:5]))

	// Create temporary file to store schema retrieved (also for cache)
	tempStorePath := filepath.Join(h.Properties.SchemaPath, strconv.Itoa(int(schemaID))+".txt")
	if !fileExists(tempStorePath) {
		schemaString, _, err := h.SrClient.DefaultApi.GetSchema(h.Ctx, schemaID, nil)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(tempStorePath, []byte(schemaString.Schema), 0644)
		if err != nil {
			return "", err
		}
	}
	return tempStorePath, nil
}

// saramaConf converts KafkaClusterConfig to sarama.Config
func saramaConf(kafka *v1.KafkaClusterConfig, clientID string, beginning bool) (*sarama.Config, error) {
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

	return saramaConf, nil
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
	msg = strings.TrimRight(msg, " \t\n")
	l.logger.Log("msg", msg)
}
