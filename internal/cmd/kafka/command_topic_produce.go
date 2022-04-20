package kafka

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/serdes"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func newProduceCommand(prerunner pcmd.PreRunner, clientId string) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "produce <topic>",
		Short:       "Produce messages to a Kafka topic.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &hasAPIKeyTopicCommand{
		HasAPIKeyCLICommand: pcmd.NewHasAPIKeyCLICommand(cmd, prerunner),
		prerunner:           prerunner,
		clientID:            clientId,
	}
	cmd.RunE = pcmd.NewCLIRunE(c.produce)

	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().String("value-format", "string", "Format of message value as string, avro, protobuf, or jsonschema. Note that schema references are not supported for avro.")
	cmd.Flags().String("refs", "", "The path to the references file.")
	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client.`)
	cmd.Flags().String("config-file", "", "The path to the configuration file.")
	cmd.Flags().String("sr-endpoint", "", "Endpoint for Schema Registry cluster.")
	cmd.Flags().String("sr-api-key", "", "Schema registry API key.")
	cmd.Flags().String("sr-api-secret", "", "Schema registry API key secret.")
	cmd.Flags().String("api-key", "", "API key.")
	cmd.Flags().String("api-secret", "", "API key secret.")
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	cmd.Flags().String("environment", "", "Environment ID.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *hasAPIKeyTopicCommand) produce(cmd *cobra.Command, args []string) error {
	topic := args[0]
	cluster, err := c.Config.Context().GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("config-file") && cmd.Flags().Changed("config") {
		return errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "config-file", "config")
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	producer, err := newProducer(cluster, c.clientID, configFile, config)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateProducerMsg, err)
	}
	defer producer.Close()
	log.CliLogger.Tracef("Create producer succeeded")

	adminClient, err := ckafka.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientMsg, err)
	}
	defer adminClient.Close()

	err = c.validateTopic(adminClient, topic, cluster)
	if err != nil {
		return err
	}

	valueFormat, subject, serializationProvider, err := prepareSerializer(cmd, topic)
	if err != nil {
		return err
	}

	schemaPath, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}
	refs, err := sr.ReadSchemaRefs(cmd)
	if err != nil {
		return err
	}

	dir, err := sr.CreateTempDir()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	// Meta info contains a magic byte and schema ID (4 bytes).
	metaInfo, referencePathMap, err := c.registerSchema(cmd, dir, valueFormat, schemaPath, subject, serializationProvider.GetSchemaName(), refs)
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
	// CCloud Kafka messageMaxBytes:
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

	// Trap SIGINT to trigger a shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	go func() {
		<-signals
		close(input)
	}()
	// Prime reader
	go scan()

	deliveryChan := make(chan ckafka.Event)
	for data := range input {
		if len(data) == 0 {
			go scan()
			continue
		}

		msg, err := getProduceMessage(cmd, metaInfo, topic, data, serializationProvider)
		if err != nil {
			return err
		}
		err = producer.Produce(msg, deliveryChan)
		if err != nil {
			isProduceToCompactedTopicError, err := errors.CatchProduceToCompactedTopicError(err, topic)
			if isProduceToCompactedTopicError {
				scanErr = err
				close(input)
				break
			}
			utils.ErrPrintf(cmd, errors.FailedToProduceErrorMsg, msg.TopicPartition.Offset, err)
		}

		e := <-deliveryChan                // read a ckafka event from the channel
		m := e.(*ckafka.Message)           // extract the message from the event
		if m.TopicPartition.Error != nil { // catch all other errors
			utils.ErrPrintf(cmd, errors.FailedToProduceErrorMsg, m.TopicPartition.Offset, m.TopicPartition.Error)
		}
		go scan()
	}
	close(deliveryChan)
	return scanErr
}

func prepareSerializer(cmd *cobra.Command, topicName string) (string, string, serdes.SerializationProvider, error) {
	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return "", "", nil, err
	}
	subject := topicNameStrategy(topicName)
	serializationProvider, err := serdes.GetSerializationProvider(valueFormat)
	if err != nil {
		return "", "", nil, err
	}
	return valueFormat, subject, serializationProvider, nil
}

func getProduceMessage(cmd *cobra.Command, metaInfo []byte, topicName, data string, serializationProvider serdes.SerializationProvider) (*ckafka.Message, error) {
	parseKey, err := cmd.Flags().GetBool("parse-key")
	if err != nil {
		return nil, err
	}
	delimiter, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return nil, err
	}
	key, value, err := getMsgKeyAndValue(metaInfo, data, delimiter, parseKey, serializationProvider)
	if err != nil {
		return nil, err
	}

	return &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &topicName, Partition: ckafka.PartitionAny},
		Key:            []byte(key),
		Value:          []byte(value),
	}, nil
}

func getMsgKeyAndValue(metaInfo []byte, data, delimiter string, parseKey bool, serializationProvider serdes.SerializationProvider) (string, string, error) {
	var key, valueString string
	if parseKey {
		record := strings.SplitN(data, delimiter, 2)
		valueString = strings.TrimSpace(record[len(record)-1])

		if len(record) == 2 {
			key = strings.TrimSpace(record[0])
		} else {
			return "", "", errors.New(errors.MissingKeyErrorMsg)
		}
	} else {
		valueString = strings.TrimSpace(data)
	}
	encodedMessage, err := serdes.Serialize(serializationProvider, valueString)
	if err != nil {
		return "", "", err
	}
	encoded := append(metaInfo, encodedMessage...)
	value := string(encoded)
	return key, value, nil
}

func (c *hasAPIKeyTopicCommand) registerSchema(cmd *cobra.Command, schemaDir, valueFormat, schemaPath, subject, schemaType string, refs []srsdk.SchemaReference) ([]byte, map[string]string, error) {
	// For plain string encoding, meta info is empty.
	// Registering schema when specified, and fill metaInfo array.
	var metaInfo []byte
	referencePathMap := map[string]string{}
	if valueFormat != "string" && len(schemaPath) > 0 {
		srAPIKey, err := cmd.Flags().GetString("sr-api-key")
		if err != nil {
			return nil, nil, err
		}
		srAPISecret, err := cmd.Flags().GetString("sr-api-secret")
		if err != nil {
			return nil, nil, err
		}

		srClient, ctx, err := sr.GetAPIClientWithAPIKey(cmd, nil, c.Config, c.Version, srAPIKey, srAPISecret)
		if err != nil {
			if err.Error() == errors.NotLoggedInErrorMsg {
				return nil, nil, new(errors.SRNotAuthenticatedError)
			} else {
				return nil, nil, err
			}
		}

		info, err := sr.RegisterSchemaWithAuth(cmd, subject, schemaType, schemaPath, refs, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}
		metaInfo = info
		referencePathMap, err = sr.StoreSchemaReferences(schemaDir, refs, srClient, ctx)
		if err != nil {
			return metaInfo, nil, err
		}
	}
	return metaInfo, referencePathMap, nil
}
