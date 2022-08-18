package asyncapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	ckgo "github.com/confluentinc/confluent-kafka-go/kafka"
	schemaregistry "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/swaggest/go-asyncapi/reflector/asyncapi-2.1.0"
	"github.com/swaggest/go-asyncapi/spec-2.1.0"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/serdes"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedStateFlagCommand
}

type TopicConfigs struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Configs struct {
	CleanupPolicy                  string `json:"cleanup.policy"`
	DeleteRetentionMs              int    `json:"delete.retention.ms"`
	ConfluentValueSchemaValidation string `json:"confluent.value.schema.validation"`
}

type ConfluentBinding struct {
	Partitions int     `json:"x-partitions"`
	Replicas   int     `json:"x-replicas"`
	Configs    Configs `json:"x-configs"`
}

type OperationBinding struct {
	GroupId  string `json:"groupId"`
	ClientId string `json:"clientId"`
}

type Key struct {
	Type string `json:"type"`
}

type MessageBinding struct {
	Key            interface{} `json:"key"`
	BindingVersion string      `json:"bindingVersion"`
}

type bindings struct {
	ChannelBindings  interface{}
	MessageBinding   interface{}
	OperationBinding interface{}
}

type SecurityConfigsSR struct {
	BasicAuthInfo string `json:"basic.auth.user.info:"`
}

type flags struct {
	file            string
	groupId         string
	consumeExamples bool
	apiKey          string
	apiSecret       string
	valueFormat     string
}

// messageOffset is 5, as the schema ID is stored at the [1:5] bytes of a message as meta info (when valid)
const messageOffset int = 5

type channelDetails struct {
	currentTopic       *schedv1.TopicDescription
	currentSubject     string
	contentType        string
	schema             *schemaregistry.Schema
	unmarshalledSchema map[string]interface{}
	mapOfMessageCompat map[string]interface{}
	tags               []spec.Tag
	bindings           *bindings
	example            interface{}
}

type accountDetails struct {
	cluster        *schedv1.KafkaCluster
	topics         []*schedv1.TopicDescription
	clusterCreds   *v1.APIKeyPair
	consumer       *ckgo.Consumer
	broker         string
	srCluster      *v1.SchemaRegistryCluster
	srClient       *schemaregistry.APIClient
	srContext      context.Context
	subjects       []string
	channelDetails channelDetails
}

func newExportCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Create an AsyncAPI specification for a Kafka cluster.",
	}
	c := &command{AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.RunE = c.export
	c.Flags().String("file", "asyncapi-spec.yaml", "Output file name.")
	c.Flags().String("group-id", "consumerApplication", "Group ID for Kafka binding.")
	c.Flags().Bool("consume-examples", false, "Consume messages from topics for populating examples.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddValueFormatFlag(cmd)
	return c.Command
}

func (c *command) export(cmd *cobra.Command, _ []string) (err error) {
	flags, err := getFlags(cmd)
	if err != nil {
		return err
	}
	accountDetails := new(accountDetails)
	err = c.getAccountDetails(accountDetails, flags)
	if err != nil {
		return err
	}
	// Servers & Info Section
	reflector := addServer(accountDetails.broker, accountDetails.srCluster)
	log.CliLogger.Debug("Generating AsyncAPI specification")
	messages := make(map[string]spec.Message)
	for _, topic := range accountDetails.topics {
		for _, subject := range accountDetails.subjects {
			if subject != topic.Name+"-value" || strings.HasPrefix(topic.Name, "_") {
				// Avoid internal topics or if subject does not follow topic naming strategy
				continue
			} else {
				// Subject and Topic matches
				accountDetails.channelDetails.currentTopic = topic
				accountDetails.channelDetails.currentSubject = subject
				err := c.getChannelDetails(accountDetails, flags)
				if err != nil {
					return err
				}
				messages[strcase.ToCamel(topic.Name)+"Message"] = spec.Message{
					OneOf1: &spec.MessageOneOf1{MessageEntity: buildMessageEntity(accountDetails)},
				}
				reflector, err = addChannel(reflector, accountDetails.channelDetails.currentTopic.Name, *accountDetails.channelDetails.bindings, accountDetails.channelDetails.mapOfMessageCompat)
				if err != nil {
					return err
				}
			}
		}
	}
	// Components
	reflector = addComponents(reflector, messages)
	// Convert reflector to YAML File
	yaml, err := reflector.Schema.MarshalYAML()
	if err != nil {
		return err
	}
	utils.Printf(cmd, "AsyncAPI specification written to \"%s\".\n", flags.file)
	return ioutil.WriteFile(flags.file, yaml, 0644)
}

func (c *command) getChannelDetails(details *accountDetails, flags *flags) error {
	err := getSchemaDetails(details)
	if details.channelDetails.contentType == "PROTOBUF" {
		return nil
	}
	if err != nil {
		return err
	}
	err = getTags(details)
	if err != nil {
		log.CliLogger.Warnf("failed to get tags: %v", err)
	}
	details.channelDetails.example = nil
	if flags.consumeExamples {
		details.channelDetails.example, err = c.getMessageExamples(details.consumer, details.channelDetails.currentTopic.Name, details.channelDetails.contentType, details.srClient, flags.valueFormat)
		if err != nil {
			log.CliLogger.Warn(err)
		}
	}
	details.channelDetails.bindings, err = c.getBindings(details.cluster, details.channelDetails.currentTopic, flags.groupId)
	if err != nil {
		return fmt.Errorf("bindings not found: %v", err)
	}
	// x-messageCompatibility
	details.channelDetails.mapOfMessageCompat, err = getMessageCompatibility(details.srClient, details.srContext, details.channelDetails.currentSubject)
	if err != nil {
		return fmt.Errorf("failed to get subject's compatibility type")
	}
	return nil
}

func (c *command) getAccountDetails(details *accountDetails, flags *flags) error {
	err := c.getClusterDetails(details)
	if err != nil {
		return err
	}
	details.broker = getBroker(details.cluster)
	err = c.getSchemaRegistry(details, flags)
	if err != nil {
		return err
	}
	details.subjects, _, err = details.srClient.DefaultApi.List(details.srContext, nil)
	if err != nil {
		return err
	}
	// Create Consumer
	if flags.consumeExamples {
		details.consumer, err = createConsumer(details.broker, details.clusterCreds, flags.groupId)
		if err != nil {
			return err
		}
		defer details.consumer.Close()
	}
	return nil
}

func getTags(details *accountDetails) error {
	tags, _, err := details.srClient.DefaultApi.GetTags(details.srContext, "sr_schema", strconv.Itoa(int(details.channelDetails.schema.Id)))
	if err != nil {
		return fmt.Errorf("failed to get schema level tags: %v", err)
	}
	var tagsInSpec []spec.Tag
	for _, tag := range tags {
		tagsInSpec = append(tagsInSpec, spec.Tag{Name: tag.TypeName})
	}
	details.channelDetails.tags = tagsInSpec
	return nil
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

func (c command) getMessageExamples(consumer *ckgo.Consumer, topicName, contentType string, srClient *schemaregistry.APIClient, valueFormatFlag string) (interface{}, error) {
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
		Ctx:        *new(context.Context),
		Format:     valueFormat,
		Out:        nil,
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

func (c *command) getBindings(cluster *schedv1.KafkaCluster, topicDescription *schedv1.TopicDescription, groupId string) (*bindings, error) {
	topic := schedv1.Topic{Spec: &schedv1.TopicSpecification{Name: topicDescription.Name}}
	configs, err := c.Client.Kafka.ListTopicConfig(context.Background(), cluster, &topic)
	if err != nil {
		return nil, fmt.Errorf("failed to get topic configs: %v", err)
	}
	var cleanupPolicy string
	deleteRetentionMsValue := -1
	for _, config := range configs.Entries {
		if config.Name == "cleanup.policy" {
			cleanupPolicy = config.Value
		}
		if config.Name == "delete.retention.ms" {
			deleteRetentionMsValue, err = strconv.Atoi(config.Value)
			if err != nil {
				return nil, err
			}
		}
	}
	channelBinding := ConfluentBinding{
		Partitions: len(topicDescription.GetPartitions()),
		Replicas:   len(topicDescription.GetPartitions()[0].Replicas),
		Configs: Configs{
			CleanupPolicy:                  cleanupPolicy,
			DeleteRetentionMs:              deleteRetentionMsValue,
			ConfluentValueSchemaValidation: "true",
		},
	}
	bindings := &bindings{
		MessageBinding: MessageBinding{
			Key:            Key{Type: "string"},
			BindingVersion: "0.1.0",
		},
		OperationBinding: OperationBinding{
			GroupId:  groupId,
			ClientId: "client1",
		},
	}
	if deleteRetentionMsValue != -1 && cleanupPolicy != "" {
		bindings.ChannelBindings = channelBinding
	}
	return bindings, nil
}

func (c *command) getClusterDetails(details *accountDetails) error {
	var ctx context.Context
	kafkaClusterId := c.Config.Context().KafkaClusterContext.GetActiveKafkaClusterId()
	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId(), Id: kafkaClusterId}
	// Get Kafka Cluster
	cluster, err := c.Client.Kafka.Describe(ctx, req)
	if err != nil {
		return fmt.Errorf(`failed to describe cluster: %v`, err)
	}
	clusterConfig, err := c.Config.Context().FindKafkaCluster(kafkaClusterId)
	if err != nil {
		return fmt.Errorf(`failed to find Kafka cluster: %v`, err)
	}
	clusterCreds := clusterConfig.APIKeys[clusterConfig.APIKey]
	if clusterCreds == nil {
		return errors.NewErrorWithSuggestions("API key not set for the Kafka cluster", "Set an API key pair for the Kafka cluster using `confluent api-key create`")
	}
	topics, err := c.Client.Kafka.ListTopics(context.Background(), cluster)
	if err != nil {
		return fmt.Errorf("failed to get topics: %v", err)
	}
	details.cluster = cluster
	details.topics = topics
	details.clusterCreds = clusterCreds
	return nil
}

func getBroker(cluster *schedv1.KafkaCluster) string {
	return strings.Split(cluster.GetEndpoint(), "//")[1]
}

func getSchemaDetails(details *accountDetails) error {
	log.CliLogger.Debugf("Adding operation: %s", details.channelDetails.currentTopic.Name)
	schema, _, err := details.srClient.DefaultApi.GetSchemaByVersion(details.srContext, details.channelDetails.currentSubject, "latest", nil)
	if err != nil {
		return err
	}
	var unmarshalledSchema map[string]interface{}
	if schema.SchemaType == "" {
		details.channelDetails.contentType = "application/avro"
	} else if schema.SchemaType == "JSON" {
		details.channelDetails.contentType = "application/json"
	} else if schema.SchemaType == "PROTOBUF" {
		log.CliLogger.Warn("Protobuf not supported.")
		details.channelDetails.contentType = "PROTOBUF"
		return nil
	}
	// JSON or Avro Format
	err = json.Unmarshal([]byte(schema.Schema), &unmarshalledSchema)
	if err != nil {
		return fmt.Errorf("failed to unmarshal schema: %v", err)

	}
	details.channelDetails.unmarshalledSchema = unmarshalledSchema
	details.channelDetails.schema = &schema
	return nil
}

func getEnv(broker string) string {
	if strings.Contains(broker, "devel") {
		return "dev"
	} else if strings.Contains(broker, "local") {
		return "local"
	} else {
		return "prod"
	}
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
	apiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return nil, err
	}
	apiSecret, err := cmd.Flags().GetString("api-secret")
	if err != nil {
		return nil, err
	}
	valueFormat, err := cmd.Flags().GetString("value-format")
	if err != nil {
		return nil, err
	}
	return &flags{
		file:            file,
		groupId:         groupId,
		consumeExamples: consumeExamples,
		apiKey:          apiKey,
		apiSecret:       apiSecret,
		valueFormat:     valueFormat,
	}, nil
}

func (c *command) getSchemaRegistry(details *accountDetails, flags *flags) error {
	schemaCluster, err := c.Config.Context().SchemaRegistryCluster(c.Command)
	if err != nil {
		return fmt.Errorf("unable to get Schema Registry cluster: %v", err)
	}
	if flags.apiKey == "" && flags.apiSecret == "" && schemaCluster.SrCredentials != nil {
		flags.apiKey = schemaCluster.SrCredentials.Key
		flags.apiSecret = schemaCluster.SrCredentials.Secret
	}
	srClient, ctx, err := sr.GetSchemaRegistryClientWithApiKey(c.Command, c.Config, c.Version, flags.apiKey, flags.apiSecret)
	if err != nil {
		return err
	}
	details.srCluster = schemaCluster
	details.srClient = srClient
	details.srContext = ctx
	return nil
}

func addServer(broker string, schemaCluster *v1.SchemaRegistryCluster) asyncapi.Reflector {
	return asyncapi.Reflector{
		Schema: &spec.AsyncAPI{
			Servers: map[string]spec.Server{
				getEnv(broker) + "-broker": {
					URL:             broker,
					Description:     "Confluent Kafka instance.",
					ProtocolVersion: "2.6.0",
					Protocol:        "kafka",
					Security: []map[string][]string{
						{
							"confluentBroker": []string{},
						},
					},
				},
				getEnv(broker) + "-schemaRegistry": {
					URL:             schemaCluster.SchemaRegistryEndpoint,
					Description:     "Confluent Kafka Schema Registry Server",
					ProtocolVersion: "2.6.0",
					Protocol:        "kafka",
					Security: []map[string][]string{
						{
							"confluentSchemaRegistry": []string{},
						},
					},
				},
			},
			Info: spec.Info{
				Version: "1.0.0",
				Title:   "API Document for Confluent Cluster",
			},
		},
	}
}

func getMessageCompatibility(srClient *schemaregistry.APIClient, ctx context.Context, subject string) (map[string]interface{}, error) {
	var config schemaregistry.Config
	mapOfMessageCompat := make(map[string]interface{})
	config, _, err := srClient.DefaultApi.GetSubjectLevelConfig(ctx, subject, nil)
	if err != nil {
		log.CliLogger.Warnf("failed to get subject level configuration: %v", err)
		config, _, err = srClient.DefaultApi.GetTopLevelConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get top level configuration: %v", err)
		}
	}
	mapOfMessageCompat["x-messageCompatibility"] = interface{}(config.CompatibilityLevel)
	return mapOfMessageCompat, nil
}

func buildMessageEntity(details *accountDetails) *spec.MessageEntity {
	entityProducer := new(spec.MessageEntity)
	(*spec.MessageEntity).WithContentType(entityProducer, details.channelDetails.contentType)
	if details.channelDetails.contentType == "application/avro" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/vnd.apache.avro;version=1.9.0")
	} else if details.channelDetails.contentType == "application/json" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/schema+json;version=draft-07")
	}
	(*spec.MessageEntity).WithTags(entityProducer, details.channelDetails.tags...)
	// Name
	(*spec.MessageEntity).WithName(entityProducer, strcase.ToCamel(details.channelDetails.currentTopic.Name)+"Message")
	// Example
	if details.channelDetails.example != nil {
		(*spec.MessageEntity).WithExamples(entityProducer, spec.MessageOneOf1OneOf1ExamplesItems{Payload: &details.channelDetails.example})
	}
	(*spec.MessageEntity).WithBindings(entityProducer, spec.MessageBindingsObject{Kafka: &details.channelDetails.bindings.MessageBinding})
	(*spec.MessageEntity).WithPayload(entityProducer, details.channelDetails.unmarshalledSchema)
	return entityProducer
}

func addChannel(reflector asyncapi.Reflector, topicName string, bindings bindings, mapOfMessageCompat map[string]interface{}) (asyncapi.Reflector, error) {
	channel := asyncapi.ChannelInfo{
		Name: topicName,
		BaseChannelItem: &spec.ChannelItem{
			MapOfAnything: mapOfMessageCompat,
			Subscribe: &spec.Operation{
				ID:       strcase.ToCamel(topicName) + "Subscribe",
				Message:  &spec.Message{Reference: &spec.Reference{Ref: "#/components/messages/" + strcase.ToCamel(topicName) + "Message"}},
				Bindings: &spec.OperationBindingsObject{Kafka: &bindings.OperationBinding},
			},
		},
	}
	if bindings.ChannelBindings != nil {
		channel.BaseChannelItem.Bindings = &spec.ChannelBindingsObject{Kafka: &bindings.ChannelBindings}
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
							MapOfAnything: map[string]interface{}{
								"x-configs": interface{}(map[string]string{
									"basic.auth.user.info": "{{SCHEMA_REGISTRY_API_KEY}}:{{SCHEMA_REGISTRY_API_SECRET}}",
								}),
							},
						},
					},
				},
				"confluentBroker": {
					SecurityScheme: &spec.SecurityScheme{
						UserPassword: &spec.UserPassword{
							MapOfAnything: map[string]interface{}{
								"x-configs": interface{}(map[string]string{
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
