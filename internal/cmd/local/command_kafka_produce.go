package local

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/serdes"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/cobra"
)

func (c *localCommand) newProduceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "produce",
		Short: "Produce messages to the test Kafka topic.",
		Args:  cobra.NoArgs,
		RunE:  c.produce,
	}

	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client.`)
	cmd.Flags().String("config-file", "", "The path to the configuration file (in json or avro format) for the producer client.")
	return cmd
}

func (c *localCommand) produce(cmd *cobra.Command, args []string) error {
	producer, err := newOnPremProducer(cmd, ":"+c.Config.LocalPorts.PlaintextPort)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Errorf(errors.FailedToCreateProducerErrorMsg, err).Error(), errors.OnPremConfigGuideSuggestions)
	}
	defer producer.Close()
	log.CliLogger.Tracef("Create producer succeeded")

	serializationProvider, err := serdes.GetSerializationProvider("string")
	if err != nil {
		return err
	}

	output.ErrPrintln(errors.StartingProducerMsg)

	// Line reader for producer input.
	scanner := bufio.NewScanner(os.Stdin)
	// On-prem Kafka messageMaxBytes: using the same value of cloud. TODO: allow larger sizes if customers request
	// https://github.com/confluentinc/cc-spec-kafka/blob/9f0af828d20e9339aeab6991f32d8355eb3f0776/plugins/kafka/kafka.go#L43.
	const maxScanTokenSize = 1024*1024*2 + 12
	scanner.Buffer(nil, maxScanTokenSize)
	input := make(chan string, 1)
	// Avoid blocking in for loop so ^C or ^D can exit immediately.
	var scanErr error
	scan := func() {
		hasNext := scanner.Scan()
		if !hasNext {
			// Actual error.
			if scanner.Err() != nil {
				scanErr = scanner.Err()
			}
			// Otherwise just EOF.
			close(input)
		} else {
			input <- scanner.Text()
		}
	}

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

		msg, err := kafka.GetProduceMessage(cmd, []byte{0, 0, 0, 0}, testTopicName, data, serializationProvider)
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
			isProduceToCompactedTopicError, err := errors.CatchProduceToCompactedTopicError(err, testTopicName)
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
