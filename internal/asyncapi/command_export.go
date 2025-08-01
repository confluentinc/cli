package asyncapi

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/swaggest/go-asyncapi/reflector/asyncapi-2.4.0"
	"github.com/swaggest/go-asyncapi/spec-2.4.0"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/confluentinc/cli/v4/internal/kafka"
	"github.com/confluentinc/cli/v4/pkg/auth"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	pkafka "github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/schemaregistry"
	"github.com/confluentinc/cli/v4/pkg/serdes"
)

type confluentBinding struct {
	BindingVersion     string                   `json:"bindingVersion,omitempty"`
	Partitions         int32                    `json:"partitions,omitempty"`
	TopicConfiguration topicConfigurationExport `json:"topicConfiguration,omitempty"`
	XConfigs           map[string]string        `json:"x-configs,omitempty"`
}

type topicConfigurationExport struct {
	CleanupPolicy       []string `json:"cleanup.policy,omitempty"`
	RetentionTime       int64    `json:"retention.ms,omitempty"`
	RetentionSize       int64    `json:"retention.bytes,omitempty"`
	DeleteRetentionTime int64    `json:"delete.retention.ms,omitempty"`
	MaxMessageSize      int32    `json:"max.message.bytes,omitempty"`
}

type bindings struct {
	channelBindings  spec.ChannelBindingsObject
	messageBinding   spec.MessageBindingsObject
	operationBinding spec.OperationBindingsObject
}

type flags struct {
	file            string
	group           string
	consumeExamples bool
	specVersion     string
	kafkaApiKey     string
	valueFormat     string
	schemaContext   string
	topics          []string
}

const protobufErrorMessage = "protobuf is not supported"

func (c *command) newExportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export an AsyncAPI specification.",
		Long:  "Export an AsyncAPI specification for a Kafka cluster and Schema Registry.",
		Args:  cobra.NoArgs,
		RunE:  c.export,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Export an AsyncAPI specification with topic "my-topic" and all topics starting with "prefix-".`,
				Code: `confluent asyncapi export --topics "my-topic,prefix-*"`,
			},
		),
	}

	cmd.Flags().String("file", "asyncapi-spec.yaml", "Output file name.")
	cmd.Flags().String("group", "consumerApplication", "Consumer Group ID for getting messages.")
	cmd.Flags().Bool("consume-examples", false, "Consume messages from topics for populating examples.")
	cmd.Flags().String("spec-version", "1.0.0", "Version number of the output file.")
	cmd.Flags().String("kafka-api-key", "", "Kafka cluster API key.")
	cmd.Flags().String("schema-context", "default", "Use a specific schema context.")
	cmd.Flags().StringSlice("topics", nil, "A comma-separated list of topics to export. Supports prefixes ending with a wildcard (*).")
	cmd.Flags().String("schema-registry-endpoint", "", "The URL of the Schema Registry cluster.")
	pcmd.AddValueFormatFlag(cmd)
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	// Deprecated
	cmd.Flags().String("schema-registry-api-key", "", "API key for Schema Registry.")
	cobra.CheckErr(cmd.Flags().MarkHidden("schema-registry-api-key"))

	// Deprecated
	cmd.Flags().String("schema-registry-api-secret", "", "API secret for Schema Registry.")
	cobra.CheckErr(cmd.Flags().MarkHidden("schema-registry-api-secret"))

	cobra.CheckErr(cmd.MarkFlagFilename("file", "yaml", "yml"))

	return cmd
}

func (c *command) export(cmd *cobra.Command, _ []string) error {
	flags, err := getFlags(cmd)
	if err != nil {
		return err
	}
	accountDetails, err := c.getAccountDetails(cmd, flags)
	if err != nil {
		return err
	}
	if flags.consumeExamples {
		defer accountDetails.consumer.Close()
	}
	// Servers & Info Section
	reflector := addServer(accountDetails.kafkaUrl, accountDetails.schemaRegistryUrl, flags.specVersion)
	log.CliLogger.Debug("Generating AsyncAPI specification")
	messages := make(map[string]spec.Message)
	var schemaContextPrefix string
	if flags.schemaContext != "default" {
		log.CliLogger.Debugf(`Using schema context "%s"`, flags.schemaContext)
		schemaContextPrefix = fmt.Sprintf(":.%s:", flags.schemaContext)
	}
	channelCount := 0

	for _, topic := range accountDetails.topics {
		if !topicMatch(topic.GetTopicName(), flags.topics) {
			continue
		}

		for _, subject := range accountDetails.subjects {
			if subject != fmt.Sprintf("%s%s-value", schemaContextPrefix, topic.GetTopicName()) || strings.HasPrefix(topic.GetTopicName(), "_") {
				// Avoid internal topics or if subject does not follow topic naming strategy
				continue
			}
			// Subject and Topic matches
			// Reset channel details
			accountDetails.channelDetails = channelDetails{
				currentTopic:   topic,
				currentSubject: subject,
				examples:       make(map[string]any),
			}
			if err := c.getChannelDetails(accountDetails, flags, cmd); err != nil {
				if err.Error() == protobufErrorMessage {
					log.CliLogger.Info(err.Error())
					continue
				}
				return err
			}
			channelCount++
			messages[msgName(topic.GetTopicName())] = spec.Message{
				OneOf1: &spec.MessageOneOf1{MessageEntity: accountDetails.buildMessageEntity()},
			}
			reflector, err = addChannel(reflector, accountDetails.channelDetails)
			if err != nil {
				return err
			}
		}
	}
	// if no channels, add an empty object
	if channelCount == 0 {
		reflector.Schema.Channels = map[string]spec.ChannelItem{}
	}
	// Components
	reflector = addComponents(reflector, messages)
	// Convert reflector to YAML File
	yaml, err := reflector.Schema.MarshalYAML()
	if err != nil {
		return err
	}
	output.Printf(c.Config.EnableColor, "AsyncAPI specification written to \"%s\".\n", flags.file)
	return os.WriteFile(flags.file, yaml, 0644)
}

func (c *command) getChannelDetails(details *accountDetails, flags *flags, cmd *cobra.Command) error {
	topic := details.channelDetails.currentTopic.GetTopicName()
	if err := details.getSchemaDetails(); err != nil {
		if err.Error() == protobufErrorMessage {
			return err
		}
		return fmt.Errorf("failed to get schema details: %w", err)
	}
	if err := details.getTags(); err != nil {
		log.CliLogger.Warnf("Failed to get tags: %v", err)
	}
	if flags.consumeExamples {
		example, err := c.getMessageExamples(details.consumer, topic, details.channelDetails.contentType, details.srClient, flags.valueFormat, cmd)
		if err != nil {
			log.CliLogger.Warn(err)
		}
		details.channelDetails.examples[topic] = example
	}
	bindings, err := c.getBindings(cmd, topic)
	if err != nil {
		log.CliLogger.Warnf("Bindings not found: %v", err)
	}
	details.channelDetails.bindings = bindings
	if err := details.getTopicDescription(); err != nil {
		log.CliLogger.Warnf("Failed to get topic description: %v", err)
	}
	// x-messageCompatibility
	mapOfMessageCompat, err := getMessageCompatibility(details.srClient, details.channelDetails.currentSubject)
	if err != nil {
		log.CliLogger.Warnf("Failed to get subject's compatibility type: %v", err)
	}
	details.channelDetails.mapOfMessageCompat = mapOfMessageCompat
	output.Printf(c.Config.EnableColor, "Added topic \"%s\".\n", topic)
	return nil
}

func (c *command) getAccountDetails(cmd *cobra.Command, flags *flags) (*accountDetails, error) {
	details := new(accountDetails)
	if err := c.getClusterDetails(details, flags, cmd); err != nil {
		return nil, err
	}

	srClient, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return nil, err
	}
	details.srClient = srClient

	subjects, err := details.srClient.List("", false)
	if err != nil {
		return nil, err
	}
	details.subjects = subjects

	// Create Consumer
	if flags.consumeExamples {
		details.consumer, err = createConsumer(details.kafkaBootstrapUrl, details.clusterCreds, flags.group)
		if err != nil {
			return nil, err
		}
	}

	return details, nil
}

func getValueFormat(contentType string) string {
	switch contentType {
	case "application/avro":
		return "avro"
	case "application/json":
		return "jsonschema"
	case "application/protobuf":
		return "protobuf"
	default:
		return "string"
	}
}

func handlePanic() {
	if err := recover(); err != nil {
		log.CliLogger.Warnf("Failed to get message example: %v", err)
	}
}

func (c *command) getMessageExamples(consumer *ckgo.Consumer, topicName, contentType string, srClient *schemaregistry.Client, valueFormatFlag string, cmd *cobra.Command) (any, error) {
	defer handlePanic()
	if err := consumer.Subscribe(topicName, nil); err != nil {
		return nil, fmt.Errorf(`failed to subscribe to topic "%s": %w`, topicName, err)
	}
	message, err := consumer.ReadMessage(10 * time.Second)
	if err != nil {
		return nil, fmt.Errorf(`no example received for topic "%s": %w`, topicName, err)
	}
	value := message.Value
	var valueFormat string
	if valueFormatFlag != "" {
		valueFormat = valueFormatFlag
	} else {
		valueFormat = getValueFormat(contentType)
	}
	deserializationProvider, err := serdes.GetDeserializationProvider(valueFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to get deserializer for %s", valueFormat)
	}

	token, err := auth.GetDataplaneToken(c.Context)
	if err != nil {
		return nil, err
	}
	schemaRegistryEndpoint, err := cmd.Flags().GetString("schema-registry-endpoint")
	if err != nil {
		return nil, err
	}
	var srClusterId, srEndpoint string
	if schemaRegistryEndpoint != "" {
		srEndpoint = schemaRegistryEndpoint
		clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(c.Context.GetCurrentEnvironment())
		if err != nil {
			return nil, err
		}
		if len(clusters) == 0 {
			return nil, schemaregistry.ErrNotEnabled
		}
		srClusterId = clusters[0].GetId()
	} else {
		srClusterId, srEndpoint, err = c.GetCurrentSchemaRegistryClusterIdAndEndpoint(cmd)
	}
	if err != nil {
		return nil, err
	}

	err = deserializationProvider.InitDeserializer(srEndpoint, srClusterId, "value", serdes.SchemaRegistryAuth{Token: token}, nil)
	if err != nil {
		return nil, err
	}
	groupHandler := kafka.GroupHandler{
		SrClient:    srClient,
		ValueFormat: valueFormat,
		Subject:     topicName + "-value",
		Properties:  kafka.ConsumerProperties{},
	}
	if slices.Contains(serdes.SchemaBasedFormats, valueFormat) {
		schemaPath, referencePathMap, err := groupHandler.RequestSchema(value)
		if err != nil {
			return nil, err
		}
		if err := deserializationProvider.LoadSchema(schemaPath, referencePathMap); err != nil {
			return nil, err
		}
	}

	jsonMessage, err := deserializationProvider.Deserialize(topicName, value)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize example: %v", err)
	}

	var jsonObject map[string]interface{}
	if err := json.Unmarshal([]byte(jsonMessage), &jsonObject); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON message: %v", err)
	}

	if valueFormat == "avro" {
		processUnionTypes(jsonObject)
	}

	return jsonObject, nil
}

// Valid avro types
var avroTypes = map[string]bool{
	"string":  true,
	"array":   true,
	"int":     true,
	"long":    true,
	"float":   true,
	"double":  true,
	"boolean": true,
	"null":    true,
	"bytes":   true,
}

// processUnionTypes recursively processes the message map to handle union types
func processUnionTypes(m map[string]any) {
	for k, v := range m {
		switch val := v.(type) {
		case map[string]any:
			// Handle Avro union types
			if len(val) == 1 {
				for typeName, actualValue := range val {
					if avroTypes[typeName] {
						m[k] = actualValue
						continue
					}
				}
			}
			// Handle arrays
			if arrayVal, ok := val["array"]; ok {
				if array, ok := arrayVal.([]any); ok {
					for i, item := range array {
						if itemMap, ok := item.(map[string]any); ok {
							processUnionTypes(itemMap)
							array[i] = itemMap
						}
					}
					m[k] = array
					continue
				}
			}
			// Process nested maps
			processUnionTypes(val)
		case []any:
			// Handle arrays that might contain union types
			for i, item := range val {
				if itemMap, ok := item.(map[string]any); ok {
					processUnionTypes(itemMap)
					val[i] = itemMap
				}
			}
		}
	}
}

func (c *command) getBindings(cmd *cobra.Command, topicName string) (*bindings, error) {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return nil, err
	}
	configs, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(topicName)
	if err != nil {
		return nil, err
	}
	var numPartitions int32
	partitions, err := kafkaREST.CloudClient.ListKafkaPartitions(topicName)
	if err != nil {
		return nil, fmt.Errorf("unable to get topic partitions: %w", err)
	}
	if partitions != nil {
		numPartitions = int32(len(partitions))
	}
	customConfigMap := make(map[string]string)
	topicConfigMap := make(map[string]any)

	// Determine whether the given config value can be put into the AsyncAPI Kafka bindings or put into our custom struct for extra configs
	for _, config := range configs {
		switch config.GetName() {
		case "cleanup.policy":
			topicConfigMap[config.GetName()] = strings.Split(config.GetValue(), ",")
		case "max.message.bytes":
			topicConfigMap[config.GetName()], err = strconv.ParseInt(config.GetValue(), 10, 32)
			if err != nil {
				return nil, err
			}
		case "delete.retention.ms", "retention.bytes", "retention.ms":
			topicConfigMap[config.GetName()], err = strconv.ParseInt(config.GetValue(), 10, 64)
			if err != nil {
				return nil, err
			}
		default:
			customConfigMap[config.GetName()] = config.GetValue()
		}
	}

	// Turn topicConfigMap into correct format
	topicConfigs := topicConfigurationExport{}
	jsonString, err := json.Marshal(topicConfigMap)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonString, &topicConfigs); err != nil {
		return nil, err
	}

	var channelBindings any = confluentBinding{
		BindingVersion:     "0.4.0",
		Partitions:         numPartitions,
		TopicConfiguration: topicConfigs,
		XConfigs:           customConfigMap,
	}
	messageBindings := spec.MessageBindingsObject{Kafka: &spec.KafkaMessage{Key: &spec.KafkaMessageKey{Schema: map[string]any{"type": "string"}}}}
	operationBindings := spec.OperationBindingsObject{Kafka: &spec.KafkaOperation{
		GroupID:  &spec.KafkaOperationGroupID{Schema: map[string]any{"type": "string"}},
		ClientID: &spec.KafkaOperationClientID{Schema: map[string]any{"type": "string"}},
	}}
	bindings := &bindings{
		messageBinding:   messageBindings,
		operationBinding: operationBindings,
	}
	bindings.channelBindings = spec.ChannelBindingsObject{Kafka: &channelBindings}

	return bindings, nil
}

func (c *command) getClusterDetails(details *accountDetails, flags *flags, cmd *cobra.Command) error {
	cluster, err := pkafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return fmt.Errorf(`failed to find Kafka cluster: %w`, err)
	}

	if err := cluster.DecryptAPIKeys(); err != nil {
		return err
	}

	var clusterCreds *config.APIKeyPair
	if flags.kafkaApiKey != "" {
		if _, ok := cluster.APIKeys[flags.kafkaApiKey]; !ok {
			key, httpResp, err := c.V2Client.GetApiKey(flags.kafkaApiKey)
			if err != nil {
				return errors.CatchCCloudV2Error(err, httpResp)
			}
			// check if the key is for the right cluster
			if key.Spec.Resource.Id != cluster.ID {
				return errors.NewErrorWithSuggestions(
					fmt.Sprintf(errors.InvalidApiKeyErrorMsg, flags.kafkaApiKey, cluster.ID),
					fmt.Sprintf(errors.InvalidApiKeySuggestions, cluster.ID),
				)
			}
			// the requested api-key exists, but the secret is not saved locally
			return &errors.UnconfiguredAPISecretError{APIKey: flags.kafkaApiKey, ClusterID: cluster.ID}
		}
		clusterCreds = cluster.APIKeys[flags.kafkaApiKey]
	} else {
		clusterCreds = cluster.APIKeys[cluster.APIKey]
	}
	if clusterCreds == nil {
		return errors.NewErrorWithSuggestions(
			"API key not set for the Kafka cluster",
			"Set an API key pair for the Kafka cluster using `confluent api-key create --resource <cluster-id>` and then use it with `--kafka-api-key`.",
		)
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	topics, err := kafkaREST.CloudClient.ListKafkaTopics()
	if err != nil {
		return err
	}

	environment, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	clusters, err := c.V2Client.GetSchemaRegistryClustersByEnvironment(environment)
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		return schemaregistry.ErrNotEnabled
	}

	details.kafkaClusterId = cluster.ID
	details.schemaRegistryClusterId = clusters[0].GetId()
	details.clusterCreds = clusterCreds
	// kafkaUrl is used in exported file, while kafkaBootstrapUrl is used to create Kafka consumer
	details.kafkaUrl = kafkaREST.CloudClient.GetUrl()
	details.kafkaBootstrapUrl = cluster.Bootstrap
	schemaRegistryEndpoint, err := cmd.Flags().GetString("schema-registry-endpoint")
	if err != nil {
		return err
	}
	if schemaRegistryEndpoint != "" {
		details.schemaRegistryUrl = schemaRegistryEndpoint
	} else if c.Context.GetSchemaRegistryEndpoint() != "" {
		details.schemaRegistryUrl = c.Context.GetSchemaRegistryEndpoint()
	} else {
		endpoint, err := c.GetValidSchemaRegistryClusterEndpoint(cmd, clusters[0])
		if err != nil {
			return err
		}
		details.schemaRegistryUrl = endpoint
	}

	details.topics = topics.Data

	return nil
}

func getFlags(cmd *cobra.Command) (*flags, error) {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return nil, err
	}
	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}
	consumeExamples, err := cmd.Flags().GetBool("consume-examples")
	if err != nil {
		return nil, err
	}
	specVersion, err := cmd.Flags().GetString("spec-version")
	if err != nil {
		return nil, err
	}
	kafkaApiKey, err := cmd.Flags().GetString("kafka-api-key")
	if err != nil {
		return nil, err
	}
	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return nil, err
	}
	schemaContext, err := cmd.Flags().GetString("schema-context")
	if err != nil {
		return nil, err
	}
	topics, err := cmd.Flags().GetStringSlice("topics")
	if err != nil {
		return nil, err
	}
	return &flags{
		file:            file,
		group:           group,
		consumeExamples: consumeExamples,
		specVersion:     specVersion,
		kafkaApiKey:     kafkaApiKey,
		valueFormat:     valueFormat,
		schemaContext:   schemaContext,
		topics:          topics,
	}, nil
}

func msgName(s string) string {
	return strcase.ToCamel(s) + "Message"
}

func addServer(kafkaUrl, schemaRegistryUrl, specVersion string) asyncapi.Reflector {
	return asyncapi.Reflector{
		Schema: &spec.AsyncAPI{
			Servers: map[string]spec.ServersAdditionalProperties{
				"cluster": {Server: &spec.Server{
					URL:         kafkaUrl,
					Description: "Confluent Kafka instance.",
					Protocol:    "kafka",
					Security:    []map[string][]string{{"confluentBroker": []string{}}},
				}},
				"schema-registry": {Server: &spec.Server{
					URL:         schemaRegistryUrl,
					Description: "Confluent Kafka Schema Registry Server",
					Protocol:    "kafka",
					Security:    []map[string][]string{{"confluentSchemaRegistry": []string{}}},
				}},
			},
			Info: spec.Info{
				Version: specVersion,
				Title:   "API Document for Confluent Cluster",
			},
		},
	}
}

func getMessageCompatibility(client *schemaregistry.Client, subject string) (map[string]any, error) {
	config, err := client.GetSubjectLevelConfig(subject)
	if err != nil {
		log.CliLogger.Warnf("Failed to get subject level configuration: %v", err)
		config, err = client.GetTopLevelConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get top level configuration: %w", err)
		}
	}
	return map[string]any{"x-messageCompatibility": config.CompatibilityLevel}, nil
}

func addChannel(reflector asyncapi.Reflector, details channelDetails) (asyncapi.Reflector, error) {
	channel := asyncapi.ChannelInfo{
		Name: details.currentTopic.GetTopicName(),
		BaseChannelItem: &spec.ChannelItem{
			Description: details.currentTopicDescription,
			Subscribe: &spec.Operation{
				ID:   strcase.ToCamel(details.currentTopic.GetTopicName()) + "Subscribe",
				Tags: details.topicLevelTags,
			},
		},
	}
	if details.mapOfMessageCompat != nil {
		channel.BaseChannelItem.MapOfAnything = details.mapOfMessageCompat
	}
	if details.unmarshalledSchema != nil {
		channel.BaseChannelItem.Subscribe.Message = &spec.Message{Reference: &spec.Reference{Ref: "#/components/messages/" + msgName(details.currentTopic.GetTopicName())}}
	}
	if details.bindings != nil {
		if details.bindings.operationBinding.Kafka != nil {
			channel.BaseChannelItem.Subscribe.Bindings = &details.bindings.operationBinding
		}
		if details.bindings.channelBindings.Kafka != nil {
			channel.BaseChannelItem.Bindings = &details.bindings.channelBindings
		}
	}
	err := reflector.AddChannel(channel)
	return reflector, err
}

func addComponents(reflector asyncapi.Reflector, messages map[string]spec.Message) asyncapi.Reflector {
	reflector.Schema.WithComponents(spec.Components{
		Messages: messages,
		SecuritySchemes: &spec.ComponentsSecuritySchemes{
			MapOfComponentsSecuritySchemesWDValues: map[string]spec.ComponentsSecuritySchemesWD{
				"confluentSchemaRegistry": {
					SecurityScheme: &spec.SecurityScheme{
						UserPassword: &spec.UserPassword{
							MapOfAnything: map[string]any{
								"x-configs": any(map[string]string{
									"basic.auth.user.info": "{{SCHEMA_REGISTRY_API_KEY}}:{{SCHEMA_REGISTRY_API_SECRET}}",
								}),
							},
						},
					},
				},
				"confluentBroker": {
					SecurityScheme: &spec.SecurityScheme{
						UserPassword: &spec.UserPassword{
							MapOfAnything: map[string]any{
								"x-configs": any(map[string]string{
									"security.protocol": "sasl_ssl",
									"sasl.mechanisms":   "PLAIN",
									"sasl.username":     "{{CLUSTER_API_KEY}}",
									"sasl.password":     "{{CLUSTER_API_SECRET}}",
								}),
							},
						},
					},
				},
			},
		},
	})
	return reflector
}

func createConsumer(broker string, clusterCreds *config.APIKeyPair, group string) (*ckgo.Consumer, error) {
	consumer, err := ckgo.NewConsumer(&ckgo.ConfigMap{
		"bootstrap.servers":  broker,
		"sasl.mechanisms":    "PLAIN",
		"security.protocol":  "SASL_SSL",
		"sasl.username":      clusterCreds.Key,
		"sasl.password":      clusterCreds.Secret,
		"group.id":           group,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}
	return consumer, nil
}

// Check if topic matches user-specified topics/prefixes. True if user didn't specify
func topicMatch(topic string, userTopics []string) bool {
	if len(userTopics) == 0 {
		return true
	}

	for _, userTopic := range userTopics {
		if strings.HasSuffix(userTopic, "*") && strings.HasPrefix(topic, strings.TrimSuffix(userTopic, "*")) || userTopic == topic {
			return true
		}
	}
	return false
}
