package asyncapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	str "strings"
	"time"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	schemaregistry "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/swaggest/go-asyncapi/reflector/asyncapi-2.1.0"
	"github.com/swaggest/go-asyncapi/spec-2.1.0"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
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

type SecurityConfigsSR struct {
	BasicAuthInfo string `json:"basic.auth.user.info:"`
}

func newExportCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Create asyncapi spec for the Kafka cluster.",
	}
	c := &command{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	c.RunE = c.export
	c.Flags().String("file", "asyncapi-spec.yaml", "Output file name.")
	c.Flags().String("group-id", "consumerApplication", "Group ID for Kafka Binding.")
	c.Flags().Bool("consume-examples", false, "Consume Messages from Topics for populating Examples.")
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
	getExamples, err := cmd.Flags().GetBool("consume-examples")
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
	//For Getting Broker URL
	cluster, topics, clusterCreds, broker, err := getClusterDetails(c)
	if err != nil {
		return err
	}
	//Creating Consumer
	var consumer *kafka.Consumer
	if getExamples {
		consumer, err = kafka.NewConsumer(&kafka.ConfigMap{
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
			return err
		}
	}
	schemaCluster, err := c.Config.Context().SchemaRegistryCluster(cmd)
	if err != nil {
		return fmt.Errorf("unable to get schema registry cluster: %s", err)
	}

	if apiKey == "" && apiSecret == "" && schemaCluster.SrCredentials != nil {
		apiKey = schemaCluster.SrCredentials.Key
		apiSecret = schemaCluster.SrCredentials.Secret
	}

	srClient, ctx, err := sr.GetAPIClientWithAPIKey(cmd, nil, c.Config, c.Version, apiKey, apiSecret)
	if err != nil {
		return nil
	}
	//environment type - local, devel or prod
	env := getEnv(broker)
	//Servers & Info Section
	reflector, err := AddServer(env, broker, schemaCluster)
	if err != nil {
		return err
	}
	//SR Client
	subjects, _, err := srClient.DefaultApi.List(ctx, nil)
	if err != nil {
		return err
	}

	log.CliLogger.Debug("Generating AsyncAPI Spec...")

	messages := make(map[string]spec.Message)
	for idx := 0; idx < len(topics); idx++ {
		//For a given topic
		for i := 0; i < len(subjects); i++ {
			if subjects[i] != (topics[idx].Name+"-value") || str.HasPrefix(topics[idx].Name, "_") {
				//Avoid internal topics or if no schema is set for value.
				continue
			} else {
				//Subject and Topic matches
				contentType, tags, msgBindings, bindings, opBindings, producer, mapOfMessageCompat, err := getChannelDetails(topics[idx], srClient, ctx, subjects[i], schemaCluster, cluster, apiKey, apiSecret, clusterCreds, groupId)
				if contentType == "PROTOBUF" {
					continue
				}
				if err != nil {
					return err
				}
				reflector, err = AddMessage(reflector, topics[idx].Name, contentType, tags, getExamples, consumer, msgBindings, bindings, opBindings, messages, producer, mapOfMessageCompat)
				if err != nil {
					return err
				}
			}
		}
	}

	//Components
	reflector, err = AddComponents(reflector, messages)
	if err != nil {
		return err
	}
	if getExamples {
		err = consumer.Close()
		if err != nil {
			log.CliLogger.Warn("Error in Closing the Consumer.")
			return err
		}
	}
	//Convert reflector to YAML File
	yaml, err := reflector.Schema.MarshalYAML()
	if err != nil {
		return err
	}
	fmt.Printf("\nSpec generated in file %s in current folder.\n", file)
	err = ioutil.WriteFile(file, yaml, 0644)
	if err != nil {
		return err
	}
	return nil
}

func getTags(schemaCluster *v1.SchemaRegistryCluster, prodSchema schemaregistry.Schema, apiKey, apiSecret string) ([]spec.Tag, error) {
	dataCatalogUrl := schemaCluster.SchemaRegistryEndpoint + "/catalog/v1/entity/type/sr_schema/name/" + schemaCluster.Id + ":.:" + strconv.Itoa(int(prodSchema.Id)) + "/tags"
	req, _ := http.NewRequest("GET", dataCatalogUrl, nil)
	req.SetBasicAuth(apiKey, apiSecret)
	resp, _ := http.DefaultClient.Do(req)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.CliLogger.Warn("Error in getting Tags")
		}
	}(resp.Body)
	var tag []TagsFromId
	body, _ := ioutil.ReadAll(resp.Body)
	err := json.Unmarshal(body, &tag)
	if err != nil {
		return nil, err
	}
	var tags []spec.Tag
	for j := 0; j < len(tag); j++ {
		tagDefsUrl := schemaCluster.SchemaRegistryEndpoint + "/catalog/v1/types/tagdefs/" + tag[j].TypeName
		req, _ = http.NewRequest("GET", tagDefsUrl, nil)
		req.SetBasicAuth(apiKey, apiSecret)
		resp, _ = http.DefaultClient.Do(req)
		defer func(Body io.ReadCloser) {
			err = Body.Close()
			if err != nil {
				log.CliLogger.Warn("Error in getting Tag Definitions")
			}
		}(resp.Body)
		var tagDef TagDef
		body, _ = ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(body, &tagDef)
		if err != nil {
			fmt.Println("Error in Unmarshalling tags")
			return nil, err
		}
		tags = append(tags, spec.Tag{Name: tag[j].TypeName, Description: tagDef.Description})
	}
	return tags, nil
}

func getMessageExamples(consumer *kafka.Consumer, topicName string) (interface{}, error) {
	err := consumer.Subscribe(topicName, nil)
	if err != nil {
		log.CliLogger.Warn("Error in example: Subscribing to the topic")
		return nil, err
	}
	message, err := consumer.ReadMessage(10000 * time.Millisecond)
	if err != nil {
		fmt.Printf("No example received for topic %s\n", topicName)
		return nil, err
	} else {
		var example interface{}
		val := string(message.Value)
		val = val[str.IndexRune(val, '{'):]
		err = json.Unmarshal([]byte(val), &example)
		if err != nil {
			fmt.Printf("Example received for topic %s is not a valid JSON for unmarshalling.\n", topicName)
			return nil, err
		} else {
			return example, nil
		}
	}
}

func getBindings(cluster *schedv1.KafkaCluster, topic *schedv1.TopicDescription, clusterCreds *v1.APIKeyPair, groupId string) (interface{}, interface{}, interface{}, error) {
	var binding ConfluentBinding
	//Cleanup Policy
	var CleanupPolicy TopicConfigs
	var resp *http.Response
	cleanupPolicyUrl := cluster.RestEndpoint + "/kafka/v3/clusters/" + cluster.Id + "/topics/" + topic.Name + "/configs/cleanup.policy"
	req, _ := http.NewRequest("GET", cleanupPolicyUrl, nil)
	req.SetBasicAuth(clusterCreds.Key, clusterCreds.Secret)
	resp, _ = http.DefaultClient.Do(req)
	if resp == nil {
		CleanupPolicy.Name = "cleanup.policy"
		CleanupPolicy.Value = "delete"
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.CliLogger.Warn("Error in getting Bindings")
				log.CliLogger.Warn(err)
			}
		}(resp.Body)
		body, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(body, &CleanupPolicy)
		if err != nil {
			fmt.Println("Error in Unmarshalling Topic Configs: Cleanup Policy")
			return nil, nil, nil, err
		}
	}
	//DeleteRetentionMs
	var DeleteRetentionMs TopicConfigs
	deleteRetentionMsUrl := cluster.RestEndpoint + "/kafka/v3/clusters/" + cluster.Id + "/topics/" + topic.Name + "/configs/delete.retention.ms"
	req, _ = http.NewRequest("GET", deleteRetentionMsUrl, nil)
	req.SetBasicAuth(clusterCreds.Key, clusterCreds.Secret)
	resp, _ = http.DefaultClient.Do(req)
	if resp == nil {
		//for tests
		DeleteRetentionMs.Name = "delete.retention.ms"
		DeleteRetentionMs.Value = "86400000"
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.CliLogger.Warn("Error in getting Bindings")
			}
		}(resp.Body)
		body, _ := ioutil.ReadAll(resp.Body)
		err := json.Unmarshal(body, &DeleteRetentionMs)
		if err != nil {
			fmt.Println("Error in Unmarshalling Topic Configs: DeleteRetentionMs")
			return nil, nil, nil, err
		}
	}
	binding.Partitions = len(topic.GetPartitions())
	binding.Replicas = len(topic.GetPartitions()[0].Replicas)
	binding.Configs.CleanupPolicy = CleanupPolicy.Value
	binding.Configs.DeleteRetentionMs, _ = strconv.Atoi(DeleteRetentionMs.Value)
	binding.Configs.ConfluentValueSchemaValidation = "true"
	var binding2 interface{} = binding
	//OperationBindings
	var opBindings OperationBinding
	opBindings.GroupId = groupId
	opBindings.ClientId = "client1"
	var opBindings2 interface{} = opBindings
	//MessageBindings
	var msgBindings MessageBinding
	msgBindings.Key = Key{Type: "string"}
	msgBindings.BindingVersion = "0.1.0"
	var msgBindings2 interface{} = msgBindings
	return binding2, opBindings2, msgBindings2, nil
}

func getClusterDetails(c *command) (*schedv1.KafkaCluster, []*schedv1.TopicDescription, *v1.APIKeyPair, string, error) {
	var ctx context.Context
	lkc := c.Config.Context().KafkaClusterContext.GetActiveKafkaClusterId()
	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId(), Id: lkc}
	//Getting Kafka Cluster
	cluster, err := c.Client.Kafka.Describe(ctx, req)
	if err != nil {
		log.CliLogger.Warn("Error in getting cluster")
	}
	clusterConfig, err := c.Config.Context().FindKafkaCluster(cluster.GetId())
	if err != nil {
		log.CliLogger.Warn("Unable to Find Kafka Cluster")
		return nil, nil, nil, "", err
	}
	clusterCreds := clusterConfig.APIKeys[clusterConfig.APIKey]
	if clusterCreds == nil {
		fmt.Println("Set an API Key Pair for the kafka Cluster using `confluent api-key create`")
		err = errors.New("API Key not set for the Kafka Cluster")
		return nil, nil, nil, "", err
	}
	topics, err := c.Client.Kafka.ListTopics(context.Background(), cluster)
	if err != nil {
		log.CliLogger.Warn("Error in getting topics")
		return nil, nil, nil, "", err
	}
	broker := str.Split(cluster.GetEndpoint(), "//")[1]
	return cluster, topics, clusterCreds, broker, nil
}

func getChannelDetails(topic *schedv1.TopicDescription, srClient *schemaregistry.APIClient, ctx context.Context, subject string,
	schemaCluster *v1.SchemaRegistryCluster, cluster *schedv1.KafkaCluster, apiKey, apiSecret string, clusterCreds *v1.APIKeyPair, groupId string) (string, []spec.Tag, interface{}, interface{}, interface{}, map[string]interface{}, map[string]interface{}, error) {
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
		return contentType, nil, nil, nil, nil, nil, nil, nil
	} else { //JSON or Avro Format
		err := json.Unmarshal([]byte(Schema.Schema), &producer)
		if err != nil {
			log.CliLogger.Warn("Error in unmarshalling schema")
		}
	}
	tags, err := getTags(schemaCluster, Schema, apiKey, apiSecret)
	if err != nil {
		log.CliLogger.Warn(err, "Error in getting tags")
	}
	bindings, opBindings, msgBindings, err := getBindings(cluster, topic, clusterCreds, groupId)
	if err != nil {
		log.CliLogger.Warn("Bindings not found")
		return contentType, nil, nil, nil, nil, nil, nil, err
	}
	//x-messageCompatibility
	mapOfMessageCompat, err := AddMessageCompatibility(srClient, ctx, subject)
	if err != nil {
		return contentType, nil, nil, nil, nil, nil, nil, err
	}
	return contentType, tags, msgBindings, bindings, opBindings, producer, mapOfMessageCompat, nil
}
func getEnv(broker string) string {
	var env string
	if str.Contains(broker, "devel") {
		env = "dev"
	} else if str.Contains(broker, "local") {
		env = "local"
	} else {
		env = "prod"
	}
	return env
}

//Functions to Add stuff to spec file

func AddServer(env string, broker string, schemaCluster *v1.SchemaRegistryCluster) (asyncapi.Reflector, error) {
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

func AddMessageCompatibility(srClient *schemaregistry.APIClient, ctx context.Context, subject string) (map[string]interface{}, error) {
	var config schemaregistry.Config
	mapOfMessageCompat := make(map[string]interface{})
	config, _, err := srClient.DefaultApi.GetSubjectLevelConfig(ctx, subject, nil)
	if err != nil {
		log.CliLogger.Warn("Error in getting subject level configuration")
		config, _, err = srClient.DefaultApi.GetTopLevelConfig(ctx)
		if err != nil {
			log.CliLogger.Warn("Error in getting top level config")
			//If not found, set a default value
			config.CompatibilityLevel = "BACKWARD"
		}
	}
	mapOfMessageCompat["x-messageCompatibility"] = interface{}(config.CompatibilityLevel)
	return mapOfMessageCompat, nil
}

func AddMessage(reflector asyncapi.Reflector, topicName string, contentType string, tags []spec.Tag, getExamples bool, consumer *kafka.Consumer,
	msgBindings, bindings, opBindings interface{}, messages map[string]spec.Message, producer, mapOfMessageCompat map[string]interface{}) (asyncapi.Reflector, error) {
	entityProducer := new(spec.MessageEntity)
	(*spec.MessageEntity).WithContentType(entityProducer, contentType)
	if contentType == "application/avro" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/vnd.apache.avro;version=1.9.0")
	} else if contentType == "application/json" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/schema+json;version=draft-07")
	}
	(*spec.MessageEntity).WithTags(entityProducer, tags...)

	//Name
	(*spec.MessageEntity).WithName(entityProducer, strcase.ToCamel(topicName)+"Message")
	//Example
	if getExamples {
		example, _ := getMessageExamples(consumer, topicName)
		if example != nil {
			(*spec.MessageEntity).WithExamples(entityProducer, spec.MessageOneOf1OneOf1ExamplesItems{Payload: &example})
		}
	}

	(*spec.MessageEntity).WithBindings(entityProducer, spec.MessageBindingsObject{Kafka: &msgBindings})
	messages[strcase.ToCamel(topicName)+"Message"] = spec.Message{OneOf1: &spec.MessageOneOf1{MessageEntity: (*spec.MessageEntity).WithPayload(entityProducer, producer)}}

	err := reflector.AddChannel(asyncapi.ChannelInfo{
		Name: topicName,
		BaseChannelItem: &spec.ChannelItem{
			MapOfAnything: mapOfMessageCompat,
			Subscribe: &spec.Operation{
				ID:       strcase.ToCamel(topicName) + "Subscribe",
				Message:  &spec.Message{Reference: &spec.Reference{Ref: "#/components/messages/" + strcase.ToCamel(topicName) + "Message"}},
				Bindings: &spec.OperationBindingsObject{Kafka: &opBindings},
			},
			Bindings: &spec.ChannelBindingsObject{Kafka: &bindings},
		},
	})
	if err != nil {
		return reflector, err
	}
	return reflector, nil
}

func AddComponents(reflector asyncapi.Reflector, messages map[string]spec.Message) (asyncapi.Reflector, error) {
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
