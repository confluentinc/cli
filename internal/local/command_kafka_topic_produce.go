package local

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	"github.com/spf13/cobra"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/confluentinc/cli/v4/internal/kafka"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/serdes"
)

func (c *command) newKafkaTopicProduceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "produce <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kafkaTopicProduce,
		Short: "Produce messages to a Kafka topic.",
		Long:  "Produce messages to a Kafka topic. Configuration and command guide: https://docs.confluent.io/confluent-cli/current/cp-produce-consume.html.\n\nWhen using this command, you cannot modify the message header, and the message header will not be printed out.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Produce message to topic "test" providing key.`,
				Code: "confluent local kafka topic produce test --parse-key",
			},
		),
	}

	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client. For a full list, see https://docs.confluent.io/platform/current/clients/librdkafka/html/md_CONFIGURATION.html`)
	pcmd.AddProducerConfigFileFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	cmd.MarkFlagsMutuallyExclusive("config", "config-file")

	return cmd
}

func (c *command) kafkaTopicProduce(cmd *cobra.Command, args []string) error {
	if c.Config.LocalPorts == nil {
		return errors.NewErrorWithSuggestions(errors.FailedToReadPortsErrorMsg, errors.FailedToReadPortsSuggestions)
	}
	producer, err := newOnPremProducer(cmd, c.getPlaintextBootstrapServers())
	if err != nil {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(errors.FailedToCreateProducerErrorMsg, err),
			errors.OnPremConfigGuideSuggestions,
		)
	}
	defer producer.Close()
	log.CliLogger.Tracef("Create producer succeeded")

	adminClient, err := ckgo.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	topicName := args[0]
	err = kafka.ValidateTopic(adminClient, topicName)
	if err != nil {
		return err
	}

	serializationProvider, err := serdes.GetSerializationProvider("string")
	if err != nil {
		return err
	}

	return produceToTopic(cmd, []byte{}, []byte{}, topicName, serializationProvider, serializationProvider, producer)
}

func newOnPremProducer(cmd *cobra.Command, bootstrap string) (*ckgo.Producer, error) {
	configMap := &ckgo.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             "confluent-local",
		"bootstrap.servers":                     bootstrap,
		"retry.backoff.ms":                      "250",
		"request.timeout.ms":                    "10000",
		"security.protocol":                     "PLAINTEXT",
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return nil, err
	}
	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return nil, err
	}

	err = kafka.OverwriteKafkaClientConfigs(configMap, configFile, config)
	if err != nil {
		return nil, err
	}

	if err := kafka.SetProducerDebugOption(configMap); err != nil {
		return nil, err
	}

	return ckgo.NewProducer(configMap)
}

func produceToTopic(cmd *cobra.Command, keyMetaInfo []byte, valueMetaInfo []byte, topic string, keySerializer serdes.SerializationProvider, valueSerializer serdes.SerializationProvider, producer *ckgo.Producer) error {
	keys := "Ctrl-C or Ctrl-D"
	if runtime.GOOS == "windows" {
		keys = "Ctrl-C"
	}
	output.ErrPrintf(false, "Starting Kafka Producer. Use %s to exit.\n", keys)

	var scanErr error
	input, scan := kafka.PrepareInputChannel(&scanErr)

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		input <- kafka.EOF
	}()
	// Prime reader
	go scan()

	deliveryChan := make(chan ckgo.Event)
	for data := range input {
		if data == "" {
			if scanErr != nil {
				break
			}
			go scan()
			continue
		} else if data == kafka.EOF {
			break
		}

		message, err := kafka.GetProduceMessage(cmd, keyMetaInfo, valueMetaInfo, topic, data, keySerializer, valueSerializer)
		if err != nil {
			return err
		}
		if err := producer.Produce(message, deliveryChan); err != nil {
			isProduceToCompactedTopicError, err := errors.CatchProduceToCompactedTopicError(err, topic)
			if isProduceToCompactedTopicError {
				scanErr = err
				break
			}
			output.ErrPrintf(false, errors.FailedToProduceErrorMsg, message.TopicPartition.Offset, err)
		}

		e := <-deliveryChan                // read a ckafka event from the channel
		m := e.(*ckgo.Message)             // extract the message from the event
		if m.TopicPartition.Error != nil { // catch all other errors
			output.ErrPrintf(false, errors.FailedToProduceErrorMsg, m.TopicPartition.Offset, m.TopicPartition.Error)
		}
		go scan()
	}
	return scanErr
}
