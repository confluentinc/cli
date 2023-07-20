package kafka

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/antihax/optional"
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

func (c *command) newProduceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "produce <topic>",
		Short:             "Produce messages to a Kafka topic.",
		Long:              "Produce messages to a Kafka topic.\n\nWhen using this command, you cannot modify the message header, and the message header will not be printed out.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.produce,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	cmd.Flags().String("key-schema", "", "The ID or filepath of the message key schema.")
	cmd.Flags().String("schema", "", "The ID or filepath of the message value schema.")
	pcmd.AddKeyFormatFlag(cmd)
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().String("references", "", "The path to the references file.")
	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client.`)
	pcmd.AddProducerConfigFileFlag(cmd)
	cmd.Flags().String("schema-registry-endpoint", "", "Endpoint for Schema Registry cluster.")
	cmd.Flags().String("schema-registry-api-key", "", "Schema registry API key.")
	cmd.Flags().String("schema-registry-api-secret", "", "Schema registry API key secret.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	// Deprecated
	pcmd.AddOutputFlag(cmd)
	cobra.CheckErr(cmd.Flags().MarkHidden("output"))

	// Deprecated
	cmd.Flags().Int32("schema-id", 0, "The ID of the schema.")
	cobra.CheckErr(cmd.Flags().MarkHidden("schema-id"))

	cobra.CheckErr(cmd.MarkFlagFilename("references", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	cmd.MarkFlagsMutuallyExclusive("schema", "schema-id")
	cmd.MarkFlagsMutuallyExclusive("config", "config-file")

	return cmd
}

func (c *command) produce(cmd *cobra.Command, args []string) error {
	topic := args[0]

	cluster, err := c.Config.Context().GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	if err := addApiKeyToCluster(cmd, cluster); err != nil {
		return err
	}

	keySerializer, keyMetadata, err := c.initSchemaAndGetInfo(cmd, topic, "key")
	if err != nil {
		return err
	}

	valueSerializer, valueMetadata, err := c.initSchemaAndGetInfo(cmd, topic, "value")
	if err != nil {
		return err
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

		message, err := GetProduceMessage(cmd, keyMetadata, valueMetadata, topic, data, keySerializer, valueSerializer)
		if err != nil {
			return err
		}
		if err := producer.Produce(message, deliveryChan); err != nil {
			isProduceToCompactedTopicError, err := errors.CatchProduceToCompactedTopicError(err, topic)
			if isProduceToCompactedTopicError {
				scanErr = err
				close(input)
				break
			}
			output.ErrPrintf(errors.FailedToProduceErrorMsg, message.TopicPartition.Offset, err)
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

func (c *command) getSchemaRegistryClient(cmd *cobra.Command) (*srsdk.APIClient, context.Context, error) {
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

func (c *command) registerSchema(cmd *cobra.Command, schemaCfg *sr.RegisterSchemaConfigs) ([]byte, map[string]string, error) {
	// Registering schema and fill metaInfo array.
	var metaInfo []byte // Meta info contains a magic byte and schema ID (4 bytes).
	referencePathMap := map[string]string{}

	if len(schemaCfg.SchemaPath) > 0 {
		srClient, ctx, err := c.getSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		id, err := sr.RegisterSchemaWithAuth(cmd, schemaCfg, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}
		metaInfo = sr.GetMetaInfoFromSchemaId(id)

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

func GetProduceMessage(cmd *cobra.Command, keyMetadata, valueMetadata []byte, topic, data string, keySerializer, valueSerializer serdes.SerializationProvider) (*ckafka.Message, error) {
	parseKey, err := cmd.Flags().GetBool("parse-key")
	if err != nil {
		return nil, err
	}

	delimiter, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return nil, err
	}

	key, value, err := serializeMessage(keyMetadata, valueMetadata, data, delimiter, parseKey, keySerializer, valueSerializer)
	if err != nil {
		return nil, err
	}

	message := &ckafka.Message{
		TopicPartition: ckafka.TopicPartition{
			Topic:     &topic,
			Partition: ckafka.PartitionAny,
		},
		Key:   key,
		Value: value,
	}

	return message, nil
}

func serializeMessage(keyMetadata, valueMetadata []byte, data, delimiter string, parseKey bool, keySerializer, valueSerializer serdes.SerializationProvider) ([]byte, []byte, error) {
	var serializedKey, val string
	if parseKey {
		x := strings.SplitN(data, delimiter, 2)
		if len(x) != 2 {
			return nil, nil, errors.New(errors.MissingKeyErrorMsg)
		}

		out, err := serdes.Serialize(keySerializer, strings.TrimSpace(x[0]))
		if err != nil {
			return nil, nil, err
		}
		serializedKey = string(out)

		val = strings.TrimSpace(x[1])
	} else {
		val = strings.TrimSpace(data)
	}

	serializedValue, err := serdes.Serialize(valueSerializer, val)
	if err != nil {
		return nil, nil, err
	}

	return append(keyMetadata, serializedKey...), append(valueMetadata, serializedValue...), nil
}

func (c *command) initSchemaAndGetInfo(cmd *cobra.Command, topic, mode string) (serdes.SerializationProvider, []byte, error) {
	dir, err := sr.CreateTempDir()
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	subject := topicNameStrategy(topic)

	// Deprecated
	var schemaId optional.Int32
	if mode == "value" && cmd.Flags().Changed("schema-id") {
		id, err := cmd.Flags().GetInt32("schema-id")
		if err != nil {
			return nil, nil, err
		}
		schemaId = optional.NewInt32(id)
	}

	schemaFlagName := "schema"
	if mode == "key" {
		schemaFlagName = "key-schema"
	}
	schema, err := cmd.Flags().GetString(schemaFlagName)
	if err != nil {
		return nil, nil, err
	}
	if id, err := strconv.ParseInt(schema, 10, 32); err == nil {
		schemaId = optional.NewInt32(int32(id))
	}

	var format string
	referencePathMap := map[string]string{}
	metaInfo := []byte{}

	if schemaId.IsSet() {
		srClient, ctx, err := c.getSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		schemaString, err := sr.RequestSchemaWithId(schemaId.Value(), subject, srClient, ctx)
		if err != nil {
			return nil, nil, err
		}

		format, err = serdes.FormatTranslation(schemaString.SchemaType)
		if err != nil {
			return nil, nil, err
		}

		schema, referencePathMap, err = sr.SetSchemaPathRef(schemaString, dir, subject, schemaId.Value(), srClient, ctx)
		if err != nil {
			return nil, nil, err
		}

		metaInfo = sr.GetMetaInfoFromSchemaId(schemaId.Value())
	} else {
		format, err = cmd.Flags().GetString(fmt.Sprintf("%s-format", mode))
		if err != nil {
			return nil, nil, err
		}
	}

	serializationProvider, err := serdes.GetSerializationProvider(format)
	if err != nil {
		return nil, nil, err
	}

	if schema != "" && !schemaId.IsSet() {
		// read schema info from local file and register schema
		schemaCfg := &sr.RegisterSchemaConfigs{
			SchemaDir:  dir,
			SchemaPath: schema,
			Subject:    subject,
			Format:     format,
			SchemaType: serializationProvider.GetSchemaName(),
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

	if err := serializationProvider.LoadSchema(schema, referencePathMap); err != nil {
		return nil, nil, errors.NewWrapErrorWithSuggestions(err, "failed to load schema", "Specify a schema by passing a schema ID or the path to a schema file to the `--schema` flag.")
	}

	return serializationProvider, metaInfo, nil
}
