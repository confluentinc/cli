package asyncapi

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	ckgo "github.com/confluentinc/confluent-kafka-go/kafka"
	schemaregistry "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/swaggest/go-asyncapi/reflector/asyncapi-2.4.0"
	"github.com/swaggest/go-asyncapi/spec-2.4.0"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/serdes"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
}

type Configs struct {
	CleanupPolicy                  string `json:"cleanup.policy"`
	DeleteRetentionMs              int    `json:"delete.retention.ms"`
	ConfluentValueSchemaValidation string `json:"confluent.value.schema.validation"`
}

type confluentBinding struct {
	Configs Configs `json:"x-configs"`
}

type bindings struct {
	channelBindings  spec.ChannelBindingsObject
	messageBinding   spec.MessageBindingsObject
	operationBinding spec.OperationBindingsObject
}

type flags struct {
	file                    string
	groupId                 string
	consumeExamples         bool
	specVersion             string
	kafkaApiKey             string
	schemaRegistryApiKey    string
	schemaRegistryApiSecret string
	valueFormat             string
	schemaContext           string
}

// messageOffset is 5, as the schema ID is stored at the [1:5] bytes of a message as meta info (when valid)
const messageOffset int = 5

func newExportCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Create an AsyncAPI specification for a Kafka cluster.",
	}
	c := &command{AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.RunE = c.export
	c.Flags().String("file", "asyncapi-spec.yaml", "Output file name.")
	c.Flags().String("group-id", "consumerApplication", "Consumer Group ID for getting messages.")
	c.Flags().Bool("consume-examples", false, "Consume messages from topics for populating examples.")
	c.Flags().String("spec-version", "1.0.0", "Version number of the output file.")
	c.Flags().String("kafka-api-key", "", "Kafka cluster API key.")
	c.Flags().String("schema-registry-api-key", "", "API key for Schema Registry.")
	c.Flags().String("schema-registry-api-secret", "", "API secret for Schema Registry.")
	c.Flags().String("schema-context", "default", "Use a specific schema context.")
	pcmd.AddValueFormatFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	return c.Command
}

func (c *command) export(cmd *cobra.Command, _ []string) (err error) {
	flags, err := getFlags(cmd)
	if err != nil {
		return err
	}
	accountDetails, err := c.getAccountDetails(flags)
	if err != nil {
		return err
	}
	// Servers & Info Section
	reflector := addServer(accountDetails.broker, accountDetails.srCluster, flags.specVersion)
	log.CliLogger.Debug("Generating AsyncAPI specification")
	messages := make(map[string]spec.Message)
	var schemaContextPrefix string
	if flags.schemaContext != "default" {
		log.CliLogger.Debugf("Using schema context \"%s\"\n", flags.schemaContext)
		schemaContextPrefix = fmt.Sprintf(":.%s:", flags.schemaContext)
	}
	channelCount := 0
	for _, topic := range accountDetails.topics {
		for _, subject := range accountDetails.subjects {
			if subject != schemaContextPrefix+topic.GetTopicName()+"-value" || strings.HasPrefix(topic.GetTopicName(), "_") {
				// Avoid internal topics or if subject does not follow topic naming strategy
				continue
			} else {
				// Subject and Topic matches
				channelCount++
				// Reset channel details
				accountDetails.channelDetails = channelDetails{
					currentTopic:   topic,
					currentSubject: subject,
				}
				err := c.getChannelDetails(accountDetails, flags)
				if err != nil {
					return err
				}
				messages[msgName(topic.GetTopicName())] = spec.Message{
					OneOf1: &spec.MessageOneOf1{MessageEntity: accountDetails.buildMessageEntity()},
				}
				reflector, err = addChannel(reflector, accountDetails.channelDetails)
				if err != nil {
					return err
				}
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
	err = c.countAsyncApiUsage(accountDetails)
	if err != nil {
		return err
	}
	utils.Printf(cmd, "AsyncAPI specification written to \"%s\".\n", flags.file)
	return os.WriteFile(flags.file, yaml, 0644)
}

func (c *command) getChannelDetails(details *accountDetails, flags *flags) error {
	utils.Printf(c.Command, "Adding operation: %s\n", details.channelDetails.currentTopic.GetTopicName())
	err := details.getSchemaDetails()
	if details.channelDetails.contentType == "PROTOBUF" {
		log.CliLogger.Log("Protobuf is not supported.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get schema details: %v", err)
	}
	err = details.getTags()
	if err != nil {
		log.CliLogger.Warnf("failed to get tags: %v", err)
	}
	details.channelDetails.example = nil
	if flags.consumeExamples {
		details.channelDetails.example, err = c.getMessageExamples(details.consumer, details.channelDetails.currentTopic.GetTopicName(), details.channelDetails.contentType, details.srClient, flags.valueFormat)
		if err != nil {
			log.CliLogger.Warn(err)
		}
	}
	details.channelDetails.bindings, err = c.getBindings(details.cluster, details.channelDetails.currentTopic)
	if err != nil {
		return fmt.Errorf("bindings not found: %v", err)
	}
	err = details.getTopicDescription()
	if err != nil {
		log.CliLogger.Warnf("failed to get topic description: %v", err)
	}
	// x-messageCompatibility
	details.channelDetails.mapOfMessageCompat, err = getMessageCompatibility(details.srClient, details.srContext, details.channelDetails.currentSubject)
	if err != nil {
		return fmt.Errorf("failed to get subject's compatibility type")
	}
	return nil
}

func (c *command) getAccountDetails(flags *flags) (*accountDetails, error) {
	details := new(accountDetails)
	err := c.getClusterDetails(details, flags)
	if err != nil {
		return nil, err
	}
	err = c.getSchemaRegistry(details, flags)
	if err != nil {
		return nil, err
	}
	details.subjects, _, err = details.srClient.DefaultApi.List(details.srContext, nil)
	if err != nil {
		return nil, err
	}
	// Create Consumer
	if flags.consumeExamples {
		details.consumer, err = createConsumer(details.broker, details.clusterCreds, flags.groupId)
		if err != nil {
			return nil, err
		}
		defer details.consumer.Close()
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
		log.CliLogger.Warn("failed to get message example: ", err)
	}
}

func (c command) getMessageExamples(consumer *ckgo.Consumer, topicName, contentType string, srClient *schemaregistry.APIClient, valueFormatFlag string) (any, error) {
	defer handlePanic()
	err := consumer.Subscribe(topicName, nil)
	if err != nil {
		return nil, fmt.Errorf(`failed to subscribe to topic "%s": %v`, topicName, err)
	}
	message, err := consumer.ReadMessage(10 * time.Second)
	if err != nil {
		return nil, fmt.Errorf(`no example received for topic "%s": %v`, topicName, err)
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
	groupHandler := kafka.GroupHandler{
		SrClient:   srClient,
		Ctx:        context.Background(),
		Format:     valueFormat,
		Subject:    topicName + "-value",
		Properties: kafka.ConsumerProperties{},
	}
	if valueFormat != "string" {
		schemaPath, referencePathMap, err := groupHandler.RequestSchema(value)
		if err != nil {
			return nil, err
		}
		// Message body is encoded after 5 bytes of meta information.
		value = value[messageOffset:]
		err = deserializationProvider.LoadSchema(schemaPath, referencePathMap)
		if err != nil {
			return nil, err
		}
	}
	jsonMessage, err := serdes.Deserialize(deserializationProvider, value)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize example: %v", err)
	}
	return jsonMessage, nil
}

func (c *command) getBindings(cluster *ccstructs.KafkaCluster, topicDescription kafkarestv3.TopicData) (*bindings, error) {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return nil, err
	}
	configs, err := kafkaREST.CloudClient.ListKafkaTopicConfigs(cluster.GetId(), topicDescription.GetTopicName())
	if err != nil {
		return nil, err
	}
	var cleanupPolicy string
	deleteRetentionMsValue := -1
	for _, config := range configs.Data {
		if config.Name == "cleanup.policy" {
			cleanupPolicy = config.GetValue()
		}
		if config.Name == "delete.retention.ms" {
			deleteRetentionMsValue, err = strconv.Atoi(config.GetValue())
			if err != nil {
				return nil, err
			}
		}
	}
	var channelBindings any = confluentBinding{
		Configs: Configs{
			CleanupPolicy:                  cleanupPolicy,
			DeleteRetentionMs:              deleteRetentionMsValue,
			ConfluentValueSchemaValidation: "true",
		},
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
	if deleteRetentionMsValue != -1 && cleanupPolicy != "" {
		bindings.channelBindings = spec.ChannelBindingsObject{Kafka: &channelBindings}
	}
	return bindings, nil
}

func (c *command) getClusterDetails(details *accountDetails, flags *flags) error {
	cluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return fmt.Errorf(`failed to find Kafka cluster: %v`, err)
	}
	clusterConfig, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return fmt.Errorf(`failed to find Kafka cluster: %v`, err)
	}
	var clusterCreds *v1.APIKeyPair
	if flags.kafkaApiKey != "" {
		if _, ok := clusterConfig.APIKeys[flags.kafkaApiKey]; !ok {
			return c.Context.FetchAPIKeyError(flags.kafkaApiKey, clusterConfig.ID)
		}
		clusterCreds = clusterConfig.APIKeys[flags.kafkaApiKey]
	} else {
		clusterCreds = clusterConfig.APIKeys[clusterConfig.APIKey]
	}
	if clusterCreds == nil {
		return errors.NewErrorWithSuggestions("API key not set for the Kafka cluster",
			"Set an API key pair for the Kafka cluster using `confluent api-key create --resource <cluster-id>` and then use it with `--kafka-api-key`.")
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}
	topics, httpResp, err := kafkaREST.CloudClient.ListKafkaTopics(cluster.GetId())
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	details.cluster = cluster
	details.topics = topics.Data
	details.clusterCreds = clusterCreds
	details.broker = kafkaREST.CloudClient.GetUrl()
	return nil
}

func getFlags(cmd *cobra.Command) (*flags, error) {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return nil, err
	}
	groupId, err := cmd.Flags().GetString("group-id")
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
	schemaRegistryApiKey, err := cmd.Flags().GetString("schema-registry-api-key")
	if err != nil {
		return nil, err
	}
	schemaRegistryApiSecret, err := cmd.Flags().GetString("schema-registry-api-secret")
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
	return &flags{
		file:                    file,
		groupId:                 groupId,
		consumeExamples:         consumeExamples,
		specVersion:             specVersion,
		kafkaApiKey:             kafkaApiKey,
		schemaRegistryApiKey:    schemaRegistryApiKey,
		schemaRegistryApiSecret: schemaRegistryApiSecret,
		valueFormat:             valueFormat,
		schemaContext:           schemaContext,
	}, nil
}

func (c *command) getSchemaRegistry(details *accountDetails, flags *flags) error {
	schemaCluster, err := c.Config.Context().SchemaRegistryCluster(c.Command)
	if err != nil {
		if strings.Contains(err.Error(), "Schema Registry not enabled") {
			return errors.NewErrorWithSuggestions(err.Error(), "Enable Stream Governance Essential Package to access this feature.")
		}
		return fmt.Errorf("unable to get Schema Registry cluster: %v", err)
	}
	if flags.schemaRegistryApiKey == "" && flags.schemaRegistryApiSecret == "" && schemaCluster.SrCredentials != nil {
		flags.schemaRegistryApiKey = schemaCluster.SrCredentials.Key
		flags.schemaRegistryApiSecret = schemaCluster.SrCredentials.Secret
	}
	srClient, ctx, err := sr.GetSchemaRegistryClientWithApiKey(c.Command, c.Config, c.Version, flags.schemaRegistryApiKey, flags.schemaRegistryApiSecret)
	if err != nil {
		return err
	}
	details.srCluster = schemaCluster
	details.srClient = srClient
	details.srContext = ctx
	return nil
}

func msgName(s string) string {
	return strcase.ToCamel(s) + "Message"
}
func addServer(broker string, schemaCluster *v1.SchemaRegistryCluster, specVersion string) asyncapi.Reflector {
	return asyncapi.Reflector{
		Schema: &spec.AsyncAPI{
			Servers: map[string]spec.ServersAdditionalProperties{
				"cluster": {
					Server: &spec.Server{
						URL:         broker,
						Description: "Confluent Kafka instance.",
						Protocol:    "kafka",
						Security: []map[string][]string{
							{
								"confluentBroker": []string{},
							},
						},
					},
				},
				"schema-registry": {
					Server: &spec.Server{
						URL:         schemaCluster.SchemaRegistryEndpoint,
						Description: "Confluent Kafka Schema Registry Server",
						Protocol:    "kafka",
						Security: []map[string][]string{
							{
								"confluentSchemaRegistry": []string{},
							},
						},
					},
				},
			},
			Info: spec.Info{
				Version: specVersion,
				Title:   "API Document for Confluent Cluster",
			},
		},
	}
}

func getMessageCompatibility(srClient *schemaregistry.APIClient, ctx context.Context, subject string) (map[string]any, error) {
	var config schemaregistry.Config
	mapOfMessageCompat := make(map[string]any)
	config, _, err := srClient.DefaultApi.GetSubjectLevelConfig(ctx, subject, nil)
	if err != nil {
		log.CliLogger.Warnf("failed to get subject level configuration: %v", err)
		config, _, err = srClient.DefaultApi.GetTopLevelConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get top level configuration: %v", err)
		}
	}
	mapOfMessageCompat["x-messageCompatibility"] = any(config.CompatibilityLevel)
	return mapOfMessageCompat, nil
}

func addChannel(reflector asyncapi.Reflector, details channelDetails) (asyncapi.Reflector, error) {
	channel := asyncapi.ChannelInfo{
		Name: details.currentTopic.GetTopicName(),
		BaseChannelItem: &spec.ChannelItem{
			Description:   details.currentTopicDescription,
			MapOfAnything: details.mapOfMessageCompat,
			Subscribe: &spec.Operation{
				ID:       strcase.ToCamel(details.currentTopic.GetTopicName()) + "Subscribe",
				Message:  &spec.Message{Reference: &spec.Reference{Ref: "#/components/messages/" + msgName(details.currentTopic.GetTopicName())}},
				Bindings: &details.bindings.operationBinding,
				Tags:     details.topicLevelTags,
			},
		},
	}
	if details.bindings.channelBindings.Kafka != nil {
		channel.BaseChannelItem.Bindings = &details.bindings.channelBindings
	}
	err := reflector.AddChannel(channel)
	return reflector, err
}

func addComponents(reflector asyncapi.Reflector, messages map[string]spec.Message) asyncapi.Reflector {
	reflector.Schema.WithComponents(spec.Components{Messages: messages,
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

func createConsumer(broker string, clusterCreds *v1.APIKeyPair, groupId string) (*ckgo.Consumer, error) {
	consumer, err := ckgo.NewConsumer(&ckgo.ConfigMap{
		"bootstrap.servers":  broker,
		"sasl.mechanisms":    "PLAIN",
		"security.protocol":  "SASL_SSL",
		"sasl.username":      clusterCreds.Key,
		"sasl.password":      clusterCreds.Secret,
		"group.id":           groupId,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %v", err)
	}
	return consumer, nil
}
