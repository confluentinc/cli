package kafka

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/serdes"
	"github.com/confluentinc/cli/internal/pkg/utils"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"
)

func (c *authenticatedTopicCommand) newProduceCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "produce <topic>",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremProduce),
		Short: "Produce messages to a Kafka topic.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Produce message to topic "my_topic" with SASL_SSL protocol (providing username and password).`,
				Code: `confluent kafka topic produce my_topic --url https://localhost:8092/kafka --ca-cert-path ca.crt --protocol SASL_SSL --bootstrap "localhost:19091" --username user --password secret`,
			},
		),
	}
	pcmd.AddProtocolFlag(cmd)
	pcmd.AddMechanismFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().AddFlagSet(pcmd.OnPremAuthenticationSet())
	cmd.Flags().String("schema", "", "The path to the local schema file.")
	cmd.Flags().String("value-format", "string", "Format of message value as string, avro, protobuf, or jsonschema.")
	cmd.Flags().String("refs", "", "The path to the references file.")
	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *authenticatedTopicCommand) onPremProduce(cmd *cobra.Command, args []string) error {
	level := c.Config.Logger.GetLevel()

	configMap, err := getOnPremProducerConfigMap(cmd, c.clientID)
	if err != nil {
		return err
	}

	producer, err := ckafka.NewProducer(configMap)
	if err != nil {
		if level >= log.WARN {
			c.logger.Warnf(errors.FailedToCreateProducerMsg, err)
		}
		return fmt.Errorf(errors.FailedToCreateProducerMsg, err)
	}
	defer producer.Close()
	c.logger.Tracef("Create producer succeeded")

	err = refreshOAuthBearerToken(cmd, producer, c.AuthToken())
	if err != nil {
		return err
	}

	adminClient, err := ckafka.NewAdminClientFromProducer(producer)
	if err != nil {
		if level >= log.WARN {
			c.logger.Warnf(errors.FailedToCreateAdminClientMsg, err)
		}
		return fmt.Errorf(errors.FailedToCreateAdminClientMsg, err)
	}
	defer adminClient.Close()

	topicName := args[0]
	err = c.validateTopic(adminClient, topicName)
	if err != nil {
		return err
	}

	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return err
	}
	subject := topicName + "-value"
	serializationProvider, err := serdes.GetSerializationProvider(valueFormat)
	if err != nil {
		return err
	}

	schemaPath, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}
	refs, err := readSchemaRefs(cmd)
	if err != nil {
		return err
	}
	// Meta info contains magic byte and schema ID (4 bytes).
	metaInfo, referencePathMap, err := c.registerSchema(cmd, valueFormat, schemaPath, subject, serializationProvider.GetSchemaName(), refs)
	if err != nil {
		return err
	}
	err = serializationProvider.LoadSchema(schemaPath, referencePathMap)
	if err != nil {
		return err
	}

	utils.ErrPrintln(cmd, errors.StartingProducerMsg)

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

		msg, err := getProduceMessage(cmd, metaInfo, topicName, data, serializationProvider)
		if err != nil {
			return err
		}
		err = producer.Produce(msg, deliveryChan)
		if err != nil {
			utils.ErrPrintf(cmd, errors.FailedToProduceErrorMsg, msg.TopicPartition.Offset, err)
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
			utils.ErrPrintf(cmd, errors.FailedToProduceErrorMsg, m.TopicPartition.Offset, m.TopicPartition.Error)
		}
		go scan()
	}
	close(deliveryChan)
	return scanErr
}

func (c *authenticatedTopicCommand) registerSchema(cmd *cobra.Command, valueFormat, schemaPath, subject, schemaType string, refs []srsdk.SchemaReference) ([]byte, map[string]string, error) {
	// For plain string encoding, meta info is empty.
	// Registering schema when specified, and fill metaInfo array.
	metaInfo := []byte{}
	referencePathMap := map[string]string{}
	if valueFormat != "string" && len(schemaPath) > 0 {
		srClient, ctx, err := sr.GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
		if err != nil {
			return metaInfo, nil, err
		}
		info, err := registerSchemaWithAuth(cmd, subject, schemaType, schemaPath, refs, srClient, ctx)
		if err != nil {
			return metaInfo, nil, err
		}
		metaInfo = info
		referencePathMap, err = storeSchemaReferences(refs, srClient, ctx)
		if err != nil {
			return metaInfo, nil, err
		}
	}
	return metaInfo, referencePathMap, nil
}
