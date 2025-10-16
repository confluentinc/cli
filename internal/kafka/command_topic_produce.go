package kafka

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/v4/pkg/auth"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/schemaregistry"
	"github.com/confluentinc/cli/v4/pkg/serdes"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const (
	missingKeyOrValueErrorMsg     = "missing key or value in message"
	missingOrMalformedKeyErrorMsg = "missing or malformed key in message"
	invalidHeadersErrorMsg        = "invalid header format, headers should be in the format (key%svalue)"
)

func (c *command) newProduceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "produce <topic>",
		Short:             "Produce messages to a Kafka topic.",
		Long:              "Produce messages to a Kafka topic.\n\nWhen using this command, you can specify the message header using the `--headers` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.produce,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Produce to topic "my_topic" in Confluent Cloud with a Confluent Cloud API key.`,
				Code: "confluent kafka topic produce my_topic --api-key 0000000000000000 --api-secret <API_SECRET> --bootstrap SASL_SSL://pkc-12345.us-west-2.aws.confluent.cloud:9092 --value-format avro --schema test.avsc --schema-registry-endpoint https://psrc-12345.us-west-2.aws.confluent.cloud --schema-registry-api-key 0000000000000000 --schema-registry-api-secret <SCHEMA_REGISTRY_API_SECRET>",
			},
		),
	}

	cmd.Flags().String("bootstrap", "", `Kafka cluster endpoint (Confluent Cloud); or comma-separated list of broker hosts (Confluent Platform), each formatted as "host" or "host:port".`)
	cmd.Flags().String("key-schema", "", "The ID or filepath of the message key schema.")
	cmd.Flags().String("schema", "", "The ID or filepath of the message value schema.")
	pcmd.AddKeyFormatFlag(cmd)
	pcmd.AddValueFormatFlag(cmd)
	cmd.Flags().String("references", "", "The path to the message value schema references file.")
	cmd.Flags().Bool("parse-key", false, "Parse key from the message.")
	cmd.Flags().String("delimiter", ":", "The delimiter separating each key and value.")
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of configuration overrides ("key=value") for the producer client. For a full list, see https://docs.confluent.io/platform/current/clients/librdkafka/html/md_CONFIGURATION.html`)
	pcmd.AddProducerConfigFileFlag(cmd)
	cmd.Flags().String("schema-registry-endpoint", "", "Endpoint for Schema Registry cluster.")
	cmd.Flags().StringSlice("headers", nil, `A comma-separated list of headers formatted as "key:value".`)

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
	cmd.Flags().String("client-cert-path", "", "File or directory path to client certificate to authenticate the Schema Registry client.")
	cmd.Flags().String("client-key-path", "", "File or directory path to client key to authenticate the Schema Registry client.")
	cmd.MarkFlagsRequiredTogether("client-cert-path", "client-key-path")

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
	if c.Config.IsCloudLogin() {
		return c.produceCloud(cmd, args)
	}

	if !cmd.Flags().Changed("bootstrap") { // Required if the user isn't logged into Confluent Cloud
		return fmt.Errorf(errors.RequiredFlagNotSetErrorMsg, "bootstrap")
	}

	if c.Context.GetState() == nil {
		bootstrap, err := cmd.Flags().GetString("bootstrap")
		if err != nil {
			return err
		}

		if strings.Contains(bootstrap, "confluent.cloud") {
			if err := c.prepareAnonymousContext(cmd); err != nil {
				return err
			}

			return c.produceCloud(cmd, args)
		}
	}

	return c.produceOnPrem(cmd, args)
}

func (c *command) produceCloud(cmd *cobra.Command, args []string) error {
	topic := args[0]

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed("bootstrap") {
		bootstrap, err := cmd.Flags().GetString("bootstrap")
		if err != nil {
			return err
		}
		cluster.Bootstrap = bootstrap
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

	adminClient, err := ckgo.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	if err := c.validateTopic(adminClient, topic, cluster); err != nil {
		return err
	}

	return c.produceToTopic(cmd, keyMetaInfo, valueMetaInfo, topic, keySerializer, valueSerializer, producer, false)
}

func (c *command) produceOnPrem(cmd *cobra.Command, args []string) error {
	topic := args[0]

	keySerializer, keyMetaInfo, err := c.initSchemaAndGetInfoOnPrem(cmd, topic, "key")
	if err != nil {
		return err
	}

	valueSerializer, valueMetaInfo, err := c.initSchemaAndGetInfoOnPrem(cmd, topic, "value")
	if err != nil {
		return err
	}

	parseKey, err := cmd.Flags().GetBool("parse-key")
	if err != nil {
		return err
	}

	if (cmd.Flags().Changed("key-format") || cmd.Flags().Changed("key-schema")) && !parseKey {
		return fmt.Errorf("`--parse-key` must be set when `--key-format` or `--key-schema` is set")
	}

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

	if err := c.refreshOAuthBearerToken(cmd, producer, ckgo.OAuthBearerTokenRefresh{Config: oauthConfig}); err != nil {
		return err
	}

	adminClient, err := ckgo.NewAdminClientFromProducer(producer)
	if err != nil {
		return fmt.Errorf(errors.FailedToCreateAdminClientErrorMsg, err)
	}
	defer adminClient.Close()

	if err := ValidateTopic(adminClient, topic); err != nil {
		return err
	}

	return c.produceToTopic(cmd, keyMetaInfo, valueMetaInfo, topic, keySerializer, valueSerializer, producer, true)
}

func (c *command) registerSchemaOnPrem(cmd *cobra.Command, schemaCfg *schemaregistry.RegisterSchemaConfigs) ([]byte, map[string]string, error) {
	// For plain string encoding, meta info is empty.
	// Registering schema when specified, and fill metaInfo array.
	metaInfo := []byte{}
	referencePathMap := map[string]string{}
	if slices.Contains(serdes.SchemaBasedFormats, schemaCfg.Format) && schemaCfg.SchemaPath != "" {
		if c.Context.State == nil { // require log-in to use oauthbearer token
			return nil, nil, errors.NewErrorWithSuggestions(errors.NotLoggedInErrorMsg, errors.AuthTokenSuggestions)
		}

		client, err := c.GetSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		id, err := schemaregistry.RegisterSchemaWithAuth(schemaCfg, client)
		if err != nil {
			return nil, nil, err
		}

		if output.GetFormat(cmd).IsSerialized() {
			if err := output.SerializedOutput(cmd, &schemaregistry.RegisterSchemaResponse{Id: id}); err != nil {
				return nil, nil, err
			}
		} else {
			output.Printf(c.Config.EnableColor, "Successfully registered schema with ID \"%d\".\n", id)
		}

		metaInfo = getMetaInfoFromSchemaId(id)

		referencePathMap, err = schemaregistry.StoreSchemaReferences(schemaCfg.SchemaDir, schemaCfg.Refs, client)
		if err != nil {
			return nil, nil, err
		}
	}

	return metaInfo, referencePathMap, nil
}

func (c *command) registerSchema(cmd *cobra.Command, schemaCfg *schemaregistry.RegisterSchemaConfigs) ([]byte, map[string]string, error) {
	// Registering schema and fill metaInfo array.
	var metaInfo []byte // Meta info contains a magic byte and schema ID (4 bytes).
	referencePathMap := map[string]string{}

	if len(schemaCfg.SchemaPath) > 0 {
		srClient, err := c.GetSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		id, err := schemaregistry.RegisterSchemaWithAuth(schemaCfg, srClient)
		if err != nil {
			return nil, nil, err
		}

		if output.GetFormat(cmd).IsSerialized() {
			if err := output.SerializedOutput(cmd, &schemaregistry.RegisterSchemaResponse{Id: id}); err != nil {
				return nil, nil, err
			}
		} else {
			output.Printf(c.Config.EnableColor, "Successfully registered schema with ID \"%d\".\n", id)
		}

		metaInfo = getMetaInfoFromSchemaId(id)

		referencePathMap, err = schemaregistry.StoreSchemaReferences(schemaCfg.SchemaDir, schemaCfg.Refs, srClient)
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

func GetProduceMessage(cmd *cobra.Command, keyMetaInfo, valueMetaInfo []byte, topic, data string, keySerializer, valueSerializer serdes.SerializationProvider) (*ckgo.Message, error) {
	parseKey, err := cmd.Flags().GetBool("parse-key")
	if err != nil {
		return nil, err
	}

	delimiter, err := cmd.Flags().GetString("delimiter")
	if err != nil {
		return nil, err
	}

	key, value, err := serializeMessage(keyMetaInfo, valueMetaInfo, topic, data, delimiter, parseKey, keySerializer, valueSerializer)
	if err != nil {
		return nil, err
	}

	message := &ckgo.Message{
		TopicPartition: ckgo.TopicPartition{
			Topic:     &topic,
			Partition: ckgo.PartitionAny,
		},
		Key:   key,
		Value: value,
	}

	// This error is intentionally ignored because `confluent local kafka topic produce` does not define this flag
	headers, _ := cmd.Flags().GetStringSlice("headers")
	if headers != nil {
		parsedHeaders, err := parseHeaders(headers, delimiter)
		if err != nil {
			return nil, err
		}
		message.Headers = parsedHeaders
	}

	return message, nil
}

func serializeMessage(_, _ []byte, topic, data, delimiter string, parseKey bool, keySerializer, valueSerializer serdes.SerializationProvider) ([]byte, []byte, error) {
	var serializedKey []byte
	val := data
	if parseKey {
		schemaBased := keySerializer.GetSchemaName() != ""
		key, value, err := getKeyAndValue(schemaBased, data, delimiter)
		if err != nil {
			return nil, nil, err
		}

		_, serializedKey, err = keySerializer.Serialize(topic, key)
		if err != nil {
			return nil, nil, err
		}

		val = value
	}

	_, serializedValue, err := valueSerializer.Serialize(topic, val)
	if err != nil {
		return nil, nil, err
	}

	return serializedKey, serializedValue, nil
}

func getKeyAndValue(schemaBased bool, data, delimiter string) (string, string, error) {
	dataSplit := strings.Split(data, delimiter)
	if len(dataSplit) < 2 {
		return "", "", errors.New(missingKeyOrValueErrorMsg)
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

	return "", "", errors.New(missingOrMalformedKeyErrorMsg)
}

func (c *command) initSchemaAndGetInfo(cmd *cobra.Command, topic, mode string) (serdes.SerializationProvider, []byte, error) {
	schemaDir, err := createTempDir()
	if err != nil {
		return nil, nil, err
	}

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
		client, err := c.GetSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		schemaString, err := client.GetSchema(schemaId.Value(), subject)
		if err != nil {
			return nil, nil, err
		}

		format, err = serdes.FormatTranslation(schemaString.GetSchemaType())
		if err != nil {
			return nil, nil, err
		}

		schema, referencePathMap, err = setSchemaPathRef(schemaString, schemaDir, subject, schemaId.Value(), client)
		if err != nil {
			return nil, nil, err
		}

		metaInfo = getMetaInfoFromSchemaId(schemaId.Value())
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
		schemaCfg := &schemaregistry.RegisterSchemaConfigs{
			SchemaDir:  schemaDir,
			SchemaPath: schema,
			Subject:    subject,
			Format:     format,
			SchemaType: serializationProvider.GetSchemaName(),
		}

		flag := "references"
		if mode == "key" {
			flag = "key-references"
		}
		references, err := cmd.Flags().GetString(flag)
		if err != nil {
			return nil, nil, err
		}
		refs, err := schemaregistry.ReadSchemaReferences(references)
		if err != nil {
			return nil, nil, err
		}
		schemaCfg.Refs = refs

		metaInfo, referencePathMap, err = c.registerSchema(cmd, schemaCfg)
		if err != nil {
			return nil, nil, err
		}
	}

	// Fetch the SR client endpoint during schema registration
	srEndpoint, err := cmd.Flags().GetString("schema-registry-endpoint")
	if err != nil {
		return nil, nil, err
	}

	var srClusterId string
	if (schemaId.IsSet() || schema != "") && srEndpoint == "" {
		srClusterId, srEndpoint, err = c.GetCurrentSchemaRegistryClusterIdAndEndpoint(cmd)
		if err != nil {
			return nil, nil, err
		}
	}

	// Initialize the serializer with the same SR endpoint during registration
	// The associated schema ID is also required to initialize the serializer
	srApiKey, err := cmd.Flags().GetString("schema-registry-api-key")
	if err != nil {
		return nil, nil, err
	}
	srApiSecret, err := cmd.Flags().GetString("schema-registry-api-secret")
	if err != nil {
		return nil, nil, err
	}
	var parsedSchemaId = -1
	if len(metaInfo) >= 5 {
		parsedSchemaId = int(binary.BigEndian.Uint32(metaInfo[1:5]))
	}

	var token string
	if c.Config.IsCloudLogin() { // Do not get token if users are consuming from Cloud while logged out
		token, err = auth.GetDataplaneToken(c.Context)
		if err != nil {
			return nil, nil, err
		}
	}
	srAuth := serdes.SchemaRegistryAuth{
		ApiKey:    srApiKey,
		ApiSecret: srApiSecret,
		Token:     token,
	}
	err = serializationProvider.InitSerializer(srEndpoint, srClusterId, mode, parsedSchemaId, srAuth)
	if err != nil {
		return nil, nil, err
	}

	// LoadSchema() is only required for schemas with PROTOBUF format
	if err := serializationProvider.LoadSchema(schema, referencePathMap); err != nil {
		return nil, nil, errors.NewWrapErrorWithSuggestions(err, "failed to load schema", "Specify a schema by passing a schema ID or the path to a schema file to the `--schema` flag.")
	}

	return serializationProvider, metaInfo, nil
}

func (c *command) initSchemaAndGetInfoOnPrem(cmd *cobra.Command, topic, mode string) (serdes.SerializationProvider, []byte, error) {
	schemaDir, err := createTempDir()
	if err != nil {
		return nil, nil, err
	}

	subject := topicNameStrategy(topic, mode)

	// Deprecated
	var schemaId optional.Int32
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
		client, err := c.GetSchemaRegistryClient(cmd)
		if err != nil {
			return nil, nil, err
		}

		schemaString, err := client.GetSchema(schemaId.Value(), subject)
		if err != nil {
			return nil, nil, err
		}

		format, err = serdes.FormatTranslation(schemaString.GetSchemaType())
		if err != nil {
			return nil, nil, err
		}

		schema, referencePathMap, err = setSchemaPathRef(schemaString, schemaDir, subject, schemaId.Value(), client)
		if err != nil {
			return nil, nil, err
		}

		metaInfo = getMetaInfoFromSchemaId(schemaId.Value())
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
		schemaCfg := &schemaregistry.RegisterSchemaConfigs{
			SchemaDir:  schemaDir,
			SchemaPath: schema,
			Subject:    subject,
			Format:     format,
			SchemaType: serializationProvider.GetSchemaName(),
		}

		flag := "references"
		if mode == "key" {
			flag = "key-references"
		}
		references, err := cmd.Flags().GetString(flag)
		if err != nil {
			return nil, nil, err
		}
		refs, err := schemaregistry.ReadSchemaReferences(references)
		if err != nil {
			return nil, nil, err
		}
		schemaCfg.Refs = refs

		metaInfo, referencePathMap, err = c.registerSchemaOnPrem(cmd, schemaCfg)
		if err != nil {
			return nil, nil, err
		}
	}

	// Fetch the SR client endpoint during schema registration
	srEndpoint, err := cmd.Flags().GetString("schema-registry-endpoint")
	if err != nil {
		return nil, nil, err
	}

	var parsedSchemaId = -1
	if len(metaInfo) >= 5 {
		parsedSchemaId = int(binary.BigEndian.Uint32(metaInfo[1:5]))
	}

	certificateAuthorityPath, err := cmd.Flags().GetString("certificate-authority-path")
	if err != nil {
		return nil, nil, err
	}
	clientCertPath, clientKeyPath, err := pcmd.GetClientCertAndKeyPaths(cmd)
	if err != nil {
		return nil, nil, err
	}
	var token string
	if c.Config.IsOnPremLogin() {
		token = c.Config.Context().GetAuthToken()
	}
	srAuth := serdes.SchemaRegistryAuth{
		CertificateAuthorityPath: certificateAuthorityPath,
		ClientCertPath:           clientCertPath,
		ClientKeyPath:            clientKeyPath,
		Token:                    token,
	}
	err = serializationProvider.InitSerializer(srEndpoint, "", mode, parsedSchemaId, srAuth)
	if err != nil {
		return nil, nil, err
	}

	// LoadSchema() is only required for schemas with PROTOBUF format
	if err := serializationProvider.LoadSchema(schema, referencePathMap); err != nil {
		return nil, nil, errors.NewWrapErrorWithSuggestions(err, "failed to load schema", "Specify a schema by passing a schema ID or the path to a schema file to the `--schema` flag.")
	}

	return serializationProvider, metaInfo, nil
}

func getMetaInfoFromSchemaId(id int32) []byte {
	metaInfo := []byte{0x0}
	schemaIdBuffer := make([]byte, 4)
	binary.BigEndian.PutUint32(schemaIdBuffer, uint32(id))
	return append(metaInfo, schemaIdBuffer...)
}

func setSchemaPathRef(schemaString srsdk.SchemaString, dir, subject string, schemaId int32, client *schemaregistry.Client) (string, map[string]string, error) {
	// Create temporary file to store schema retrieved (also for cache). Retry if get error retrieving schema or writing temp schema file
	tempStorePath := filepath.Join(dir, fmt.Sprintf("%s-%d.txt", subject, schemaId))
	tempRefStorePath := filepath.Join(dir, fmt.Sprintf("%s-%d.ref", subject, schemaId))
	var references []srsdk.SchemaReference

	if !utils.FileExists(tempStorePath) || !utils.FileExists(tempRefStorePath) {
		// TODO: add handler for writing schema failure
		if err := os.WriteFile(tempStorePath, []byte(schemaString.GetSchema()), 0644); err != nil {
			return "", nil, err
		}

		refBytes, err := json.Marshal(schemaString.References)
		if err != nil {
			return "", nil, err
		}
		if err := os.WriteFile(tempRefStorePath, refBytes, 0644); err != nil {
			return "", nil, err
		}
		references = schemaString.GetReferences()
	} else {
		refBlob, err := os.ReadFile(tempRefStorePath)
		if err != nil {
			return "", nil, err
		}
		if err := json.Unmarshal(refBlob, &references); err != nil {
			return "", nil, err
		}
	}
	referencePathMap, err := schemaregistry.StoreSchemaReferences(dir, references, client)
	if err != nil {
		return "", nil, err
	}
	return tempStorePath, referencePathMap, nil
}

func parseHeaders(headers []string, delimiter string) ([]ckgo.Header, error) {
	kafkaHeaders := make([]ckgo.Header, len(headers))

	for i, header := range headers {
		parts := strings.SplitN(header, delimiter, 2)

		if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" {
			return nil, fmt.Errorf(invalidHeadersErrorMsg, delimiter)
		}

		kafkaHeaders[i] = ckgo.Header{
			Key:   strings.TrimSpace(parts[0]),
			Value: []byte(strings.TrimSpace(parts[1])),
		}
	}
	return kafkaHeaders, nil
}
