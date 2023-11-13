package kafka

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	sr "github.com/confluentinc/cli/v3/internal/schema-registry"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/serdes"
)

const (
	missingKeyOrValueErrorMsg     = "missing key or value in message"
	missingOrMalformedKeyErrorMsg = "missing or malformed key in message"
)

func (c *command) newProduceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "produce <topic>",
		Short:             "Produce messages to a Kafka topic.",
		Long:              "Produce messages to a Kafka topic.\n\nWhen using this command, you cannot modify the message header, and the message header will not be printed out.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.produce,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Produce to a cloud Kafka topic named "my_topic" without logging in to Confluent Cloud.`,
				Code: "confluent kafka topic produce my_topic --api-key my-key --api-secret my-secret --bootstrap SASL_SSL://pkc-abc12:9092 --value-format avro --schema test.avsc --schema-registry-endpoint https://psrc-ab123 --schema-registry-api-key my-sr-key --schema-registry-api-secret my-sr-secret",
			},
		),
	}

	cmd.Flags().String("bootstrap", "", "Bootstrap URL for Confluent Cloud Kafka cluster.")
	cmd.Flags().String("key-schema", "", "The ID or filepath of the message key schema.")
	cmd.Flags().String("schema", "", "The ID or filepath of the message value schema.")
	pcmd.AddKeyFormatFlag(cmd)
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().String("references", "", "The path to the message value schema references file.")
	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client.`)
	pcmd.AddProducerConfigFileFlag(cmd)
	cmd.Flags().String("schema-registry-endpoint", "", "Endpoint for Schema Registry cluster.")

	// cloud-only flags
	cmd.Flags().String("key-references", "", "The path to the message key schema references file.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	cmd.Flags().String("schema-registry-api-key", "", "Schema registry API key.")
	cmd.Flags().String("schema-registry-api-secret", "", "Schema registry API secret.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Int32("schema-id", 0, "The ID of the schema.") // Deprecated
	cobra.CheckErr(cmd.Flags().MarkHidden("schema-id"))

	// on-prem only flags
	cmd.Flags().AddFlagSet(pcmd.OnPremAuthenticationSet())
	pcmd.AddProtocolFlag(cmd)
	pcmd.AddMechanismFlag(cmd, c.AuthenticatedCLICommand)

	pcmd.AddOutputFlag(cmd) // Deprecated
	cobra.CheckErr(cmd.Flags().MarkHidden("output"))

	cobra.CheckErr(cmd.MarkFlagFilename("schema", "avsc", "json", "proto"))
	cobra.CheckErr(cmd.MarkFlagFilename("key-references", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("references", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("config-file", "avsc", "json"))

	cmd.MarkFlagsMutuallyExclusive("schema", "schema-id")
	cmd.MarkFlagsMutuallyExclusive("config", "config-file")

	return cmd
}

func (c *command) produce(cmd *cobra.Command, args []string) error {
	if c.Context == nil || c.Context.State == nil {
		if !cmd.Flags().Changed("bootstrap") {
			return fmt.Errorf(errors.RequiredFlagNotSetErrorMsg, "bootstrap")
		}

		if err := c.prepareAnonymousContext(cmd); err != nil {
			return err
		}
		return c.produceCloud(cmd, args)
	} else if c.Context.Config.IsCloudLogin() {
		return c.produceCloud(cmd, args)
	} else {
		if !cmd.Flags().Changed("bootstrap") {
			return fmt.Errorf(errors.RequiredFlagNotSetErrorMsg, "bootstrap")
		}

		if !cmd.Flags().Changed("ca-location") {
			return fmt.Errorf(errors.RequiredFlagNotSetErrorMsg, "ca-location")
		}

		return c.produceOnPrem(cmd, args)
	}
}

func (c *command) produceCloud(cmd *cobra.Command, args []string) error {
	topic := args[0]

	cluster, err := c.Context.GetKafkaClusterForCommand(c.V2Client)
	if err != nil {
		return err
	}

	if err := addApiKeyToCluster(cmd, cluster); err != nil {
		return err
	}

	keySerializer, keyMetaInfo, err := c.initSchemaAndGetInfo(cmd, topic, "key")
	if err != nil {
		return err
	}

	valueSerializer, valueMetaInfo, err := c.initSchemaAndGetInfo(cmd, topic, "value")
	if err != nil {
		return err
	}

	parseKey, err := cmd.Flags().GetBool("parse-key")
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("key-format") && !parseKey {
		return fmt.Errorf("`--parse-key` must be set when `key-format` is set")
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

	return ProduceToTopic(cmd, keyMetaInfo, valueMetaInfo, topic, keySerializer, valueSerializer, producer)
}

func (c *command) produceOnPrem(cmd *cobra.Command, args []string) error {
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}
	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	producer, err := newOnPremProducer(cmd, c.clientID, configFile, config)
	if err != nil {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(errors.FailedToCreateProducerErrorMsg, err),
			errors.OnPremConfigGuideSuggestions,
		)
	}
	defer producer.Close()
	log.CliLogger.Tracef("Create producer succeeded")

	if err := c.refreshOAuthBearerToken(cmd, producer); err != nil {
		return err
	}

	adminClient, err := ckafka.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	topic := args[0]
	if err := ValidateTopic(adminClient, topic); err != nil {
		return err
	}

	keyFormat, keySubject, keySerializer, err := prepareSerializer(cmd, topic, "key")
	if err != nil {
		return err
	}

	valueFormat, valueSubject, valueSerializer, err := prepareSerializer(cmd, topic, "value")
	if err != nil {
		return err
	}

	parseKey, err := cmd.Flags().GetBool("parse-key")
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("key-format") && !parseKey {
		return fmt.Errorf("`--parse-key` must be set when `key-format` is set")
	}

	keySchema, err := cmd.Flags().GetString("key-schema")
	if err != nil {
		return err
	}

	schema, err := cmd.Flags().GetString("schema")
	if err != nil {
		return err
	}

	refs, err := sr.ReadSchemaReferences(cmd, false)
	if err != nil {
		return err
	}

	dir, err := createTempDir()
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	keySchemaConfigs := &sr.RegisterSchemaConfigs{
		Subject:    keySubject,
		SchemaDir:  dir,
		SchemaType: keySerializer.GetSchemaName(),
		Format:     keyFormat,
		SchemaPath: keySchema,
		Refs:       refs,
	}
	keyMetaInfo, keyReferencePathMap, err := c.registerSchemaOnPrem(cmd, keySchemaConfigs)
	if err != nil {
		return err
	}
	if err := keySerializer.LoadSchema(keySchema, keyReferencePathMap); err != nil {
		return err
	}

	valueSchemaConfigs := &sr.RegisterSchemaConfigs{
		Subject:    valueSubject,
		SchemaDir:  dir,
		SchemaType: valueSerializer.GetSchemaName(),
		Format:     valueFormat,
		SchemaPath: schema,
		Refs:       refs,
	}
	valueMetaInfo, referencePathMap, err := c.registerSchemaOnPrem(cmd, valueSchemaConfigs)
	if err != nil {
		return err
	}
	if err := valueSerializer.LoadSchema(schema, referencePathMap); err != nil {
		return err
	}

	return ProduceToTopic(cmd, keyMetaInfo, valueMetaInfo, topic, keySerializer, valueSerializer, producer)
}

func prepareSerializer(cmd *cobra.Command, topic, mode string) (string, string, serdes.SerializationProvider, error) {
	valueFormat, err := cmd.Flags().GetString(fmt.Sprintf("%s-format", mode))
	if err != nil {
		return "", "", nil, err
	}

	serializer, err := serdes.GetSerializationProvider(valueFormat)
	if err != nil {
		return "", "", nil, err
	}

	return valueFormat, topicNameStrategy(topic, mode), serializer, nil
}

func (c *command) registerSchemaOnPrem(cmd *cobra.Command, schemaCfg *sr.RegisterSchemaConfigs) ([]byte, map[string]string, error) {
	// For plain string encoding, meta info is empty.
	// Registering schema when specified, and fill metaInfo array.
	metaInfo := []byte{}
	referencePathMap := map[string]string{}
	if slices.Contains(serdes.SchemaBasedFormats, schemaCfg.Format) && schemaCfg.SchemaPath != "" {
		if c.Context.State == nil { // require log-in to use oauthbearer token
			return nil, nil, errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestions)
		}

		srClient, err := c.GetSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		id, err := sr.RegisterSchemaWithAuth(cmd, schemaCfg, srClient)
		if err != nil {
			return nil, nil, err
		}
		metaInfo = sr.GetMetaInfoFromSchemaId(id)

		referencePathMap, err = sr.StoreSchemaReferences(schemaCfg.SchemaDir, schemaCfg.Refs, srClient)
		if err != nil {
			return nil, nil, err
		}
	}

	return metaInfo, referencePathMap, nil
}

func (c *command) registerSchema(cmd *cobra.Command, schemaCfg *sr.RegisterSchemaConfigs) ([]byte, map[string]string, error) {
	// Registering schema and fill metaInfo array.
	var metaInfo []byte // Meta info contains a magic byte and schema ID (4 bytes).
	referencePathMap := map[string]string{}

	if len(schemaCfg.SchemaPath) > 0 {
		srClient, err := c.GetSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		id, err := sr.RegisterSchemaWithAuth(cmd, schemaCfg, srClient)
		if err != nil {
			return nil, nil, err
		}
		metaInfo = sr.GetMetaInfoFromSchemaId(id)

		referencePathMap, err = sr.StoreSchemaReferences(schemaCfg.SchemaDir, schemaCfg.Refs, srClient)
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

func GetProduceMessage(cmd *cobra.Command, keyMetaInfo, valueMetaInfo []byte, topic, data string, keySerializer, valueSerializer serdes.SerializationProvider) (*ckafka.Message, error) {
	parseKey, err := cmd.Flags().GetBool("parse-key")
	if err != nil {
		return nil, err
	}

	delimiter, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return nil, err
	}

	key, value, err := serializeMessage(keyMetaInfo, valueMetaInfo, data, delimiter, parseKey, keySerializer, valueSerializer)
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

func serializeMessage(keyMetaInfo, valueMetaInfo []byte, data, delimiter string, parseKey bool, keySerializer, valueSerializer serdes.SerializationProvider) ([]byte, []byte, error) {
	var serializedKey []byte
	val := data
	if parseKey {
		schemaBased := keySerializer.GetSchemaName() != ""
		key, value, err := getKeyAndValue(schemaBased, data, delimiter)
		if err != nil {
			return nil, nil, err
		}

		serializedKey, err = keySerializer.Serialize(key)
		if err != nil {
			return nil, nil, err
		}

		val = value
	}

	serializedValue, err := valueSerializer.Serialize(val)
	if err != nil {
		return nil, nil, err
	}

	return append(keyMetaInfo, serializedKey...), append(valueMetaInfo, serializedValue...), nil
}

func getKeyAndValue(schemaBased bool, data, delimiter string) (string, string, error) {
	dataSplit := strings.Split(data, delimiter)
	if len(dataSplit) < 2 {
		return "", "", fmt.Errorf(missingKeyOrValueErrorMsg)
	}

	if !schemaBased {
		return strings.TrimSpace(dataSplit[0]), strings.TrimSpace(strings.Join(dataSplit[1:], delimiter)), nil
	}

	key := dataSplit[0]
	if json.Valid([]byte(strings.TrimSpace(key))) {
		return strings.TrimSpace(key), strings.TrimSpace(strings.Join(dataSplit[1:], delimiter)), nil
	}

	for i, substr := range dataSplit[1:] {
		key += delimiter + substr
		if json.Valid([]byte(strings.TrimSpace(key))) {
			return strings.TrimSpace(key), strings.TrimSpace(strings.Join(dataSplit[i+2:], delimiter)), nil
		}
	}

	return "", "", fmt.Errorf(missingOrMalformedKeyErrorMsg)
}

func (c *command) initSchemaAndGetInfo(cmd *cobra.Command, topic, mode string) (serdes.SerializationProvider, []byte, error) {
	schemaDir, err := createTempDir()
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = os.RemoveAll(schemaDir)
	}()

	subject := topicNameStrategy(topic, mode)

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
		srClient, err := c.GetSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		schemaString, err := sr.RequestSchemaWithId(schemaId.Value(), subject, srClient)
		if err != nil {
			return nil, nil, err
		}

		format, err = serdes.FormatTranslation(schemaString.SchemaType)
		if err != nil {
			return nil, nil, err
		}

		schema, referencePathMap, err = sr.SetSchemaPathRef(schemaString, schemaDir, subject, schemaId.Value(), srClient)
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
			SchemaDir:  schemaDir,
			SchemaPath: schema,
			Subject:    subject,
			Format:     format,
			SchemaType: serializationProvider.GetSchemaName(),
		}
		refs, err := sr.ReadSchemaReferences(cmd, mode == "key")
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
