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
	"github.com/confluentinc/confluent-kafka-go/kafka"
	schemaregistry "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/swaggest/go-asyncapi/reflector/asyncapi-2.1.0"
	"github.com/swaggest/go-asyncapi/spec-2.1.0"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pasyncapi "github.com/confluentinc/cli/internal/pkg/asyncapi"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

type TagsFromId struct {
	TypeName string `json:"typeName"`
}

type TagDef struct {
	Category    string `json:"category"`
	Name        string `json:"name"`
	Description string `json:"description"`
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

type Bindings struct {
	ChannelBindings  interface{}
	MessageBinding   interface{}
	OperationBinding interface{}
}

type SecurityConfigsSR struct {
	BasicAuthInfo string `json:"basic.auth.user.info:"`
}

func newExportCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Create an AsyncAPI specification for a Kafka cluster.",
	}
	c := &command{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	c.RunE = c.export
	c.Flags().String("file", "asyncapi-spec.yaml", "Output file name.")
	c.Flags().String("group-id", "consumerApplication", "Group ID for Kafka binding.")
	c.Flags().Bool("consume-examples", false, "Consume messages from topics for populating examples.")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	return c.Command
}

func (c *command) export(cmd *cobra.Command, _ []string) error {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}
	groupId, err := cmd.Flags().GetString("group-id")
	if err != nil {
		return err
	}
	consumeExamples, err := cmd.Flags().GetBool("consume-examples")
	if err != nil {
		return err
	}
	apiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return err
	}
	apiSecret, err := cmd.Flags().GetString("api-secret")
	if err != nil {
		return err
	}
	// Get Kafka cluster details and broker URL
	cluster, topics, clusterCreds, broker, err := getClusterDetails(c)
	if err != nil {
		return err
	}
	// Create Consumer
	var consumer *kafka.Consumer
	if consumeExamples {
		consumer, err = createConsumer(broker, clusterCreds, groupId)
		if err != nil {
			return err
		}
		err := closeConsumer(consumer)
		if err != nil {
			return err
		}
	}
	schemaCluster, err := c.Config.Context().SchemaRegistryCluster(cmd)
	if err != nil {
		return fmt.Errorf("unable to get Schema Registry cluster: %v", err)
	}
	if apiKey == "" && apiSecret == "" && schemaCluster.SrCredentials != nil {
		apiKey = schemaCluster.SrCredentials.Key
		apiSecret = schemaCluster.SrCredentials.Secret
	}

	srClient, ctx, err := sr.GetSchemaRegistryClientWithApiKey(cmd, c.Config, c.Version, apiKey, apiSecret)
	if err != nil {
		return nil
	}
	// environment type - local, devel or prod
	env := getEnv(broker)
	// Servers & Info Section
	reflector, err := addServer(env, broker, schemaCluster)
	if err != nil {
		return err
	}
	// SR Client
	subjects, _, err := srClient.DefaultApi.List(ctx, nil)
	if err != nil {
		return err
	}
	log.CliLogger.Debug("Generating AsyncAPI spec")
	messages := make(map[string]spec.Message)
	for _, topic := range topics {
		// For a given topic
		for _, subject := range subjects {
			if subject != (topic.Name+"-value") || strings.HasPrefix(topic.Name, "_") {
				// Avoid internal topics or if no schema is set for value.
				continue
			} else {
				// Subject and Topic matches
				contentType, Schema, producer, err := getChannelDetails(topic, srClient, ctx, subject)
				if contentType == "PROTOBUF" {
					continue
				}
				if err != nil {
					return err
				}
				tags, err := getTags(schemaCluster, Schema, apiKey, apiSecret)
				if err != nil {
					log.CliLogger.Warnf("failed to get tags: %v", err)
				}
				var example interface{}
				if consumeExamples {
					example, _ = getMessageExamples(consumer, topic.Name)
				}
				bindings, err := getBindings(cluster, topic, clusterCreds, groupId)
				if err != nil {
					return fmt.Errorf("bindings not found: %v", err)
				}
				// x-messageCompatibility
				mapOfMessageCompat, err := addMessageCompatibility(srClient, ctx, subject)
				if err != nil {
					return err
				}
				reflector, err = addMessage(reflector, topic.Name, contentType, tags, example, bindings, messages, producer, mapOfMessageCompat)
				if err != nil {
					return err
				}
			}
		}
	}
	// Components
	reflector, err = addComponents(reflector, messages)
	if err != nil {
		return err
	}
	// Convert reflector to YAML File
	yaml, err := reflector.Schema.MarshalYAML()
	if err != nil {
		return err
	}
	fmt.Printf("AsyncAPI specification file generated at \"%s\".\n", file)
	return ioutil.WriteFile(file, yaml, 0644)
}

func getTags(schemaCluster *v1.SchemaRegistryCluster, prodSchema schemaregistry.Schema, apiKey, apiSecret string) ([]spec.Tag, error) {
	body := pasyncapi.GetSchemaLevelTags(schemaCluster.SchemaRegistryEndpoint, schemaCluster.Id, strconv.Itoa(int(prodSchema.Id)), apiKey, apiSecret)
	var tagsFromId []TagsFromId
	err := json.Unmarshal(body, &tagsFromId)
	if err != nil {
		return nil, err
	}
	var tagsInSpec []spec.Tag
	for j := 0; j < len(tagsFromId); j++ {
		body := pasyncapi.GetTagDefinitions(schemaCluster.SchemaRegistryEndpoint, tagsFromId[j].TypeName, apiKey, apiSecret)
		var tagDef TagDef
		err = json.Unmarshal(body, &tagDef)
		if err != nil {
			fmt.Println("Error in Unmarshalling tags")
			return nil, err
		}
		tagsInSpec = append(tagsInSpec, spec.Tag{Name: tagsFromId[j].TypeName, Description: tagDef.Description})
	}
	return tagsInSpec, nil
}

func getMessageExamples(consumer *kafka.Consumer, topicName string) (interface{}, error) {
	err := consumer.Subscribe(topicName, nil)
	if err != nil {
		log.CliLogger.Warnf("Error in subscribing to the topic: \"%v\"", err)
		return nil, nil
	}
	message, err := consumer.ReadMessage(10 * time.Second)
	if err != nil {
		log.CliLogger.Warnf("No example received for topic \"%s\": %v\n", topicName, err)
		return nil, nil
	}
	var example interface{}
	val := string(message.Value)
	val = val[strings.Index(val, "{"):]
	err = json.Unmarshal([]byte(val), &example)
	if err != nil {
		log.CliLogger.Warnf("Example received for topic \"%s\" is not a valid JSON for unmarshalling: %v\n", topicName, err)
		return nil, nil
	}
	return example, nil
}

func getBindings(cluster *schedv1.KafkaCluster, topic *schedv1.TopicDescription, clusterCreds *v1.APIKeyPair, groupId string) (Bindings, error) {
	var CleanupPolicy TopicConfigs
	var bindings Bindings
	body := pasyncapi.GetClusterCleanupPolicy(cluster.RestEndpoint, cluster.Id, topic.Name, clusterCreds)
	if body == nil {
		CleanupPolicy.Name = "cleanup.policy"
		CleanupPolicy.Value = "delete"
	} else {
		err := json.Unmarshal(body, &CleanupPolicy)
		if err != nil {
			fmt.Println("Error in Unmarshalling Topic Configs: Cleanup Policy")
			return bindings, err
		}
	}
	// DeleteRetentionMs
	var DeleteRetentionMs TopicConfigs
	body = pasyncapi.GetClusterDeleteRetentionMs(cluster.RestEndpoint, cluster.Id, topic.Name, clusterCreds)
	if body == nil {
		// for tests
		DeleteRetentionMs.Name = "delete.retention.ms"
		DeleteRetentionMs.Value = "86400000"
	} else {
		err := json.Unmarshal(body, &DeleteRetentionMs)
		if err != nil {
			fmt.Println("Error in Unmarshalling Topic Configs: DeleteRetentionMs")
			return bindings, err
		}
	}
	var channelBindings ConfluentBinding
	channelBindings.Partitions = len(topic.GetPartitions())
	channelBindings.Replicas = len(topic.GetPartitions()[0].Replicas)
	channelBindings.Configs.CleanupPolicy = CleanupPolicy.Value
	channelBindings.Configs.DeleteRetentionMs, _ = strconv.Atoi(DeleteRetentionMs.Value)
	channelBindings.Configs.ConfluentValueSchemaValidation = "true"
	// OperationBindings
	var opBindings OperationBinding
	opBindings.GroupId = groupId
	opBindings.ClientId = "client1"
	// MessageBindings
	var msgBindings MessageBinding
	msgBindings.Key = Key{Type: "string"}
	msgBindings.BindingVersion = "0.1.0"
	bindings.ChannelBindings = channelBindings
	bindings.MessageBinding = msgBindings
	bindings.OperationBinding = opBindings
	return bindings, nil
}

func getClusterDetails(c *command) (*schedv1.KafkaCluster, []*schedv1.TopicDescription, *v1.APIKeyPair, string, error) {
	var ctx context.Context
	lkc := c.Config.Context().KafkaClusterContext.GetActiveKafkaClusterId()
	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId(), Id: lkc}
	// Get Kafka Cluster
	cluster, err := c.Client.Kafka.Describe(ctx, req)
	if err != nil {
		err = fmt.Errorf("error in getting cluster: \"%v\"", err)
		return nil, nil, nil, "", err
	}
	clusterConfig, err := c.Config.Context().FindKafkaCluster(cluster.GetId())
	if err != nil {
		err = fmt.Errorf("cannot find Kafka cluster: \"%v\"", err)
		return nil, nil, nil, "", err
	}
	clusterCreds := clusterConfig.APIKeys[clusterConfig.APIKey]
	if clusterCreds == nil {
		err := errors.NewErrorWithSuggestions("API Key not set for the Kafka cluster", "Set an API Key Pair for the kafka Cluster using `confluent api-key create`")
		return nil, nil, nil, "", err
	}
	topics, err := c.Client.Kafka.ListTopics(context.Background(), cluster)
	if err != nil {
		err = fmt.Errorf("error in getting topics: \"%v\"", err)
		return nil, nil, nil, "", err
	}
	broker := strings.Split(cluster.GetEndpoint(), "//")[1]
	return cluster, topics, clusterCreds, broker, nil
}

func getChannelDetails(topic *schedv1.TopicDescription, srClient *schemaregistry.APIClient, ctx context.Context, subject string) (string, schemaregistry.Schema, map[string]interface{}, error) {
	log.CliLogger.Debug("Adding Operation : " + topic.Name + "\n")
	Schema, _, _ := srClient.DefaultApi.GetSchemaByVersion(ctx, subject, "latest", nil)
	contentType := Schema.SchemaType
	if contentType == "" {
		contentType = "application/avro"
	} else if contentType == "JSON" {
		contentType = "application/json"
	}
	var producer map[string]interface{}
	if contentType == "PROTOBUF" {
		fmt.Println("Protobuf not supported")
		return contentType, Schema, nil, nil
	} else { // JSON or Avro Format
		err := json.Unmarshal([]byte(Schema.Schema), &producer)
		if err != nil {
			err = fmt.Errorf("error in unmarshalling schema: \"%v\"", err)
			return contentType, Schema, nil, err
		}
	}
	return contentType, Schema, producer, nil
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

func addServer(env string, broker string, schemaCluster *v1.SchemaRegistryCluster) (asyncapi.Reflector, error) {
	reflector := asyncapi.Reflector{
		Schema: &spec.AsyncAPI{
			Servers: map[string]spec.Server{
				env + "-broker": {
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
				env + "-schemaRegistry": {
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
	return reflector, nil
}

func addMessageCompatibility(srClient *schemaregistry.APIClient, ctx context.Context, subject string) (map[string]interface{}, error) {
	var config schemaregistry.Config
	mapOfMessageCompat := make(map[string]interface{})
	config, _, err := srClient.DefaultApi.GetSubjectLevelConfig(ctx, subject, nil)
	if err != nil {
		log.CliLogger.Warn("Error in getting subject level configuration")
		config, _, err = srClient.DefaultApi.GetTopLevelConfig(ctx)
		if err != nil {
			log.CliLogger.Warn("Error in getting top level config")
			// If not found, set a default value
			config.CompatibilityLevel = "BACKWARD"
		}
	}
	mapOfMessageCompat["x-messageCompatibility"] = interface{}(config.CompatibilityLevel)
	return mapOfMessageCompat, nil
}

func addMessage(reflector asyncapi.Reflector, topicName string, contentType string, tags []spec.Tag, example interface{},
	bindings Bindings, messages map[string]spec.Message, producer, mapOfMessageCompat map[string]interface{}) (asyncapi.Reflector, error) {
	entityProducer := new(spec.MessageEntity)
	(*spec.MessageEntity).WithContentType(entityProducer, contentType)
	if contentType == "application/avro" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/vnd.apache.avro;version=1.9.0")
	} else if contentType == "application/json" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/schema+json;version=draft-07")
	}
	(*spec.MessageEntity).WithTags(entityProducer, tags...)
	// Name
	(*spec.MessageEntity).WithName(entityProducer, strcase.ToCamel(topicName)+"Message")
	// Example
	if example != nil {
		(*spec.MessageEntity).WithExamples(entityProducer, spec.MessageOneOf1OneOf1ExamplesItems{Payload: &example})
	}
	(*spec.MessageEntity).WithBindings(entityProducer, spec.MessageBindingsObject{Kafka: &bindings.MessageBinding})
	messages[strcase.ToCamel(topicName)+"Message"] = spec.Message{OneOf1: &spec.MessageOneOf1{MessageEntity: (*spec.MessageEntity).WithPayload(entityProducer, producer)}}
	err := reflector.AddChannel(asyncapi.ChannelInfo{
		Name: topicName,
		BaseChannelItem: &spec.ChannelItem{
			MapOfAnything: mapOfMessageCompat,
			Subscribe: &spec.Operation{
				ID:       strcase.ToCamel(topicName) + "Subscribe",
				Message:  &spec.Message{Reference: &spec.Reference{Ref: "#/components/messages/" + strcase.ToCamel(topicName) + "Message"}},
				Bindings: &spec.OperationBindingsObject{Kafka: &bindings.OperationBinding},
			},
			Bindings: &spec.ChannelBindingsObject{Kafka: &bindings.ChannelBindings},
		},
	})
	if err != nil {
		return reflector, err
	}
	return reflector, nil
}

func addComponents(reflector asyncapi.Reflector, messages map[string]spec.Message) (asyncapi.Reflector, error) {
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
								},
								),
							},
						},
					},
				},
			},
		},
	})
	return reflector, nil
}

func createConsumer(broker string, clusterCreds *v1.APIKeyPair, groupId string) (*kafka.Consumer, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
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
		log.CliLogger.Warn("Error in creating Kafka Consumer")
		return nil, err
	}
	return consumer, nil
}

func closeConsumer(consumer *kafka.Consumer) (err error) {
	defer func(consumer *kafka.Consumer) {
		err = consumer.Close()
		err = fmt.Errorf("failed to close consumer: %v", err)
	}(consumer)
	return err
}
