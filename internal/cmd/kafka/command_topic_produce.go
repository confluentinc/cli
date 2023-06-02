package kafka

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/serdes"
)

func newProduceCommand(prerunner pcmd.PreRunner, clientId string) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "produce <topic>",
		Short:       "Produce messages to a Kafka topic.",
		Long:        "Produce messages to a Kafka topic.\n\nWhen using this command, you cannot modify the message header, and the message header will not be printed out.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &hasAPIKeyTopicCommand{
		HasAPIKeyCLICommand: pcmd.NewHasAPIKeyCLICommand(cmd, prerunner),
		prerunner:           prerunner,
		clientID:            clientId,
	}
	cmd.RunE = c.produce

	cmd.Flags().String("schema", "", "The path to the schema file.")
	cmd.Flags().Int32("schema-id", 0, "The ID of the schema.")
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().String("references", "", "The path to the references file.")
	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client.`)
	cmd.Flags().String("config-file", "", "The path to the configuration file (in JSON or avro format) for the producer client.")
	cmd.Flags().String("schema-registry-endpoint", "", "Endpoint for Schema Registry cluster.")
	cmd.Flags().String("schema-registry-api-key", "", "Schema registry API key.")
	cmd.Flags().String("schema-registry-api-secret", "", "Schema registry API key secret.")
	cmd.Flags().String("api-key", "", "API key.")
	cmd.Flags().String("api-secret", "", "API key secret.")
	cmd.Flags().String("cluster", "", "Kafka cluster ID.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	cmd.Flags().String("environment", "", "Environment ID.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagFilename("schema", "avsc", "json", "proto"))
	cobra.CheckErr(cmd.MarkFlagFilename("references", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	return cmd
}

func (c *hasAPIKeyTopicCommand) produce(cmd *cobra.Command, args []string) error {
	topic := args[0]
	cluster, err := c.Config.Context().GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("schema") && cmd.Flags().Changed("schema-id") {
		return errors.Errorf(errors.ProhibitedFlagCombinationErrorMsg, "schema", "schema-id")
	}

	serializationProvider, metaInfo, err := c.initSchemaAndGetInfo(cmd, topic)
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
		return fmt.Errorf(errors.FailedToCreateProducerErrorMsg, err)
	}
	defer producer.Close()
	log.CliLogger.Tracef("Create producer succeeded")

	adminClient, err := ckafka.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	if err := c.validateTopic(adminClient, topic, cluster); err != nil {
		return err
	}

	output.ErrPrintln(errors.StartingProducerMsg)

	var scanErr error
	input, scan := PrepareInputChannel(&scanErr)

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

		msg, err := GetProduceMessage(cmd, metaInfo, topic, data, serializationProvider)
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
			output.ErrPrintf(errors.FailedToProduceErrorMsg, msg.TopicPartition.Offset, err)
		}

		e := <-deliveryChan                // read a ckafka event from the channel
		m := e.(*ckafka.Message)           // extract the message from the event
		if m.TopicPartition.Error != nil { // catch all other errors
			output.ErrPrintf(errors.FailedToProduceErrorMsg, m.TopicPartition.Offset, m.TopicPartition.Error)
		}
		go scan()
	}
	close(deliveryChan)
	return scanErr
}

func (c *hasAPIKeyTopicCommand) getSchemaRegistryClient(cmd *cobra.Command) (*srsdk.APIClient, context.Context, error) {
	schemaRegistryApiKey, err := cmd.Flags().GetString("schema-registry-api-key")
	if err != nil {
		return nil, nil, err
	}
	schemaRegistryApiSecret, err := cmd.Flags().GetString("schema-registry-api-secret")
	if err != nil {
		return nil, nil, err
	}

	srClient, ctx, err := sr.GetSchemaRegistryClientWithApiKey(cmd, c.Config, c.Version, schemaRegistryApiKey, schemaRegistryApiSecret)
	if err != nil && err.Error() == errors.NotLoggedInErrorMsg {
		err = new(errors.SRNotAuthenticatedError)
	}
	return srClient, ctx, err
}

func (c *hasAPIKeyTopicCommand) registerSchema(cmd *cobra.Command, schemaCfg *sr.RegisterSchemaConfigs) ([]byte, map[string]string, error) {
	// Registering schema and fill metaInfo array.
	var metaInfo []byte // Meta info contains a magic byte and schema ID (4 bytes).
	referencePathMap := map[string]string{}

	if len(*schemaCfg.SchemaPath) > 0 {
		srClient, ctx, err := c.getSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		info, err := sr.RegisterSchemaWithAuth(cmd, schemaCfg, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}
		metaInfo = info
		referencePathMap, err = sr.StoreSchemaReferences(schemaCfg.SchemaDir, schemaCfg.Refs, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}
	}
	return metaInfo, referencePathMap, nil
}

func PrepareInputChannel(scanErr *error) (chan string, func()) {
	// Line reader for producer input.
	scanner := bufio.NewScanner(os.Stdin)
	// On-prem Kafka messageMaxBytes: using the same value of cloud. TODO: allow larger sizes if customers request
	// https://github.com/confluentinc/cc-spec-kafka/blob/9f0af828d20e9339aeab6991f32d8355eb3f0776/plugins/kafka/kafka.go#L43.
	const maxScanTokenSize = 1024*1024*2 + 12
	scanner.Buffer(nil, maxScanTokenSize)
	input := make(chan string, 1)
	// Avoid blocking in for loop so ^C or ^D can exit immediately.
	return input, func() {
		hasNext := scanner.Scan()
		if !hasNext {
			// Actual error.
			if scanner.Err() != nil {
				*scanErr = scanner.Err()
			}
			// Otherwise just EOF.
			close(input)
		} else {
			input <- scanner.Text()
		}
	}
}

func GetProduceMessage(cmd *cobra.Command, metaInfo []byte, topicName, data string, serializationProvider serdes.SerializationProvider) (*ckafka.Message, error) {
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
	value := string(append(metaInfo, encodedMessage...))
	return key, value, nil
}

func (c *hasAPIKeyTopicCommand) initSchemaAndGetInfo(cmd *cobra.Command, topic string) (serdes.SerializationProvider, []byte, error) {
	dir, err := sr.CreateTempDir()
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	subject := topicNameStrategy(topic)

	schemaId, err := cmd.Flags().GetInt32("schema-id")
	if err != nil {
		return nil, nil, err
	}

	schemaPath, err := cmd.Flags().GetString("schema")
	if err != nil {
		return nil, nil, err
	}

	var valueFormat string
	referencePathMap := map[string]string{}
	metaInfo := []byte{}

	if cmd.Flags().Changed("schema-id") {
		// request schema information from schemaID
		srClient, ctx, err := c.getSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		schemaString, err := sr.RequestSchemaWithId(schemaId, subject, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}

		valueFormat, err = serdes.FormatTranslation(schemaString.SchemaType)
		if err != nil {
			return nil, nil, err
		}

		schemaPath, referencePathMap, err = sr.SetSchemaPathRef(schemaString, dir, subject, schemaId, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}

		metaInfo = sr.GetMetaInfoFromSchemaId(schemaId)
	} else {
		valueFormat, err = cmd.Flags().GetString("value-format")
		if err != nil {
			return nil, nil, err
		}
	}

	serializationProvider, err := serdes.GetSerializationProvider(valueFormat)
	if err != nil {
		return nil, nil, err
	}

	if schemaPath != "" && !cmd.Flags().Changed("schema-id") {
		// read schema info from local file and register schema
		schemaCfg := &sr.RegisterSchemaConfigs{
			SchemaDir:   dir,
			SchemaPath:  &schemaPath,
			Subject:     subject,
			ValueFormat: valueFormat,
			SchemaType:  serializationProvider.GetSchemaName(),
		}
		refs, err := sr.ReadSchemaRefs(cmd)
		if err != nil {
			return nil, nil, err
		}
		schemaCfg.Refs = refs
		metaInfo, referencePathMap, err = c.registerSchema(cmd, schemaCfg)
		if err != nil {
			return nil, nil, err
		}
	}

	err = serializationProvider.LoadSchema(schemaPath, referencePathMap)
	if err != nil {
		return nil, nil, errors.NewWrapErrorWithSuggestions(err, "failed to load schema", errors.FailedToLoadSchemaSuggestions)
	}

	return serializationProvider, metaInfo, nil
}
