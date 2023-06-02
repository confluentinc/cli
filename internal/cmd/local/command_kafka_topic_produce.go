package local

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/serdes"
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
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client.`)
	cmd.Flags().String("config-file", "", "The path to the configuration file (in JSON or avro format) for the producer client.")

	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	return cmd
}

func (c *command) kafkaTopicProduce(cmd *cobra.Command, args []string) error {
	if c.Config.LocalPorts == nil {
		return errors.NewErrorWithSuggestions(errors.FailedToReadPortsErrorMsg, errors.FailedToReadPortsSuggestions)
	}
	producer, err := newOnPremProducer(cmd, ":"+c.Config.LocalPorts.PlaintextPort)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateProducerErrorMsg, err).Error(), errors.OnPremConfigGuideSuggestions)
	}
	defer producer.Close()
	log.CliLogger.Tracef("Create producer succeeded")

	adminClient, err := ckafka.NewAdminClientFromProducer(producer)
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

	output.ErrPrintln(errors.StartingProducerMsg)

	var scanErr error
	input, scan := kafka.PrepareInputChannel(&scanErr)

	signals := make(chan os.Signal, 1) // Trap SIGINT to trigger a shutdown.
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		close(input)
	}()
	go scan() // Prime reader

	deliveryChan := make(chan ckafka.Event)
	for data := range input {
		if len(data) == 0 {
			go scan()
			continue
		}

		msg, err := kafka.GetProduceMessage(cmd, make([]byte, 4), topicName, data, serializationProvider)
		if err != nil {
			return err
		}
		err = producer.Produce(msg, deliveryChan)
		if err != nil {
			output.ErrPrintf(errors.FailedToProduceErrorMsg, msg.TopicPartition.Offset, err)
		}

		e := <-deliveryChan                // read a ckafka event from the channel
		m := e.(*ckafka.Message)           // extract the message from the event
		if m.TopicPartition.Error != nil { // catch all other errors
			isProduceToCompactedTopicError, err := errors.CatchProduceToCompactedTopicError(err, topicName)
			if isProduceToCompactedTopicError {
				scanErr = err
				close(input)
				break
			}
			output.ErrPrintf(errors.FailedToProduceErrorMsg, m.TopicPartition.Offset, m.TopicPartition.Error)
		}
		go scan()
	}
	close(deliveryChan)
	return scanErr
}

func newOnPremProducer(cmd *cobra.Command, bootstrap string) (*ckafka.Producer, error) {
	configMap := &ckafka.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             "confluent-local",
		"bootstrap.servers":                     bootstrap,
		"retry.backoff.ms":                      "250",
		"request.timeout.ms":                    "10000",
		"security.protocol":                     "PLAINTEXT",
	}

	if cmd.Flags().Changed("config-file") && cmd.Flags().Changed("config") {
		return nil, errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "config-file", "config")
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

	return ckafka.NewProducer(configMap)
}
