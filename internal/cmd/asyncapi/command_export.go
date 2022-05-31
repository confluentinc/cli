package asyncapi

import (
	"context"
	"encoding/json"
	"fmt"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	schemaregistry "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/swaggest/go-asyncapi/reflector/asyncapi-2.1.0"
	"github.com/swaggest/go-asyncapi/spec-2.1.0"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	str "strings"
	"time"
)

type createCommand struct {
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
	//GroupId    string  `json:"group-id"`
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
		Short: "Used to create a yaml file by fetching info from cloud.",
	}
	c := &createCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	c.RunE = c.export
	c.Flags().StringP("file", "f", "asyncapi-spec.yaml", "Output file name.")
	c.Flags().StringP("group-id", "g", "consumerApplication", "Group ID for Kafka Binding.")
	c.Flags().Bool("consume-examples", false, "Consume Messages from Topics for populating Examples.")
	//c.Flags().String("srKey", "", "API Key for Schema Registry")
	//c.Flags().String("srSecret", "", "API Secret for Schema Registry")
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	return c.Command
}

func (c *createCommand) export(cmd *cobra.Command, _ []string) error {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		log.CliLogger.Warn("Error in File flag")
		return err
	}
	groupId, err := cmd.Flags().GetString("group-id")
	if err != nil {
		log.CliLogger.Warn("Error in GroupID flag")
	}
	getExamples, _ := cmd.Flags().GetBool("consume-examples")
	srKey, _ := cmd.Flags().GetString("api-key")
	srSecret, _ := cmd.Flags().GetString("api-secret")

	//For Getting Broker URL
	cluster, topics, clusterCreds, err := getClusterDetails(c)
	if err != nil {
		return err
	}
	broker := str.Split(cluster.GetEndpoint(), "//")[1]
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
		log.CliLogger.Warn("Unable to get API Client")
		return err
	}

	if srKey == "" && srSecret == "" && schemaCluster.SrCredentials != nil {
		srKey = schemaCluster.SrCredentials.Key
		srSecret = schemaCluster.SrCredentials.Secret
	}
	//fmt.Println(schemaCluster.SrCredentials.Key)
	//fmt.Println(schemaCluster.Id)

	//fmt.Println(hasKey)
	//SR Client
	//srClient, ctx, err := sr.GetApiClient(cmd, nil, c.Config, c.Version)
	srClient, ctx, err := sr.GetAPIClientWithAPIKey(cmd, nil, c.Config, c.Version, srKey, srSecret)

	if err != nil {
		return nil
	}
	//env
	var env string
	if str.Contains(broker, "devel") {
		env = "dev"
	} else if str.Contains(broker, "local") {
		env = "local"
	} else {
		env = "prod"
	}
	//Servers & Info Section
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

	mustNotFail := func(err error) {
		if err != nil {
			panic(err.Error())
		}
	}

	//SR Client
	subjects, _, err := srClient.DefaultApi.List(ctx, nil)
	if err != nil {
		log.CliLogger.Warn("Can't get list of subjects")
		return err
	}

	fmt.Println("Generating AsyncAPI Spec...")

	messages := make(map[string]spec.Message)
	for idx := 0; idx < len(topics); idx++ {
		//For a given topic
		for i := 0; i < len(subjects); i++ {
			if subjects[i] != (topics[idx].Name+"-value") || str.HasPrefix(topics[idx].Name, "_") {
				continue
			} else {
				//Subject and Topic matches
				fmt.Printf("Adding Operation : %s\n", topics[idx].Name)
				prodSchema, _, _ := srClient.DefaultApi.GetSchemaByVersion(ctx, subjects[i], "latest", nil)
				contentType := prodSchema.SchemaType
				if contentType == "" {
					contentType = "application/avro"
				} else if contentType == "JSON" {
					contentType = "application/json"
				}
				var producer map[string]interface{}
				//var producer2 Schema
				if contentType == "PROTOBUF" {
					fmt.Println("Protobuf not supported")
					continue
				} else { //JSON or Avro Format
					err = json.Unmarshal([]byte(prodSchema.Schema), &producer)
					if err != nil {
						log.CliLogger.Warn("Error in unmarshalling schema")
					}
				}

				//Creating Message Entity
				entityProducer := new(spec.MessageEntity)
				(*spec.MessageEntity).WithContentType(entityProducer, contentType)

				tags1, err := getTags(schemaCluster, prodSchema, srKey, srSecret)
				if err != nil {
					log.CliLogger.Warn(err, "Error in getting tags")
				}
				(*spec.MessageEntity).WithTags(entityProducer, tags1...)

				//Name
				(*spec.MessageEntity).WithName(entityProducer, strcase.ToCamel(topics[idx].Name)+"Message")
				//Example
				if getExamples {
					example, _ := getMessageExamples(consumer, topics[idx].Name)
					if example != nil {
						(*spec.MessageEntity).WithExamples(entityProducer, spec.MessageOneOf1OneOf1ExamplesItems{Payload: &example})
					}
				}

				//Schema Format
				if contentType == "application/avro" {
					(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/vnd.apache.avro;version=1.9.0")
				} else if contentType == "application/json" {
					(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/schema+json;version=draft-07")

				}

				//x-messageCompatibility
				var config schemaregistry.Config
				mapOfMessageCompat := make(map[string]interface{})
				config, _, err = srClient.DefaultApi.GetSubjectLevelConfig(ctx, subjects[i], nil)
				if err != nil {
					log.CliLogger.Warn("Error in getting subject level configuration")
					//fmt.Println(err)
					config, _, err = srClient.DefaultApi.GetTopLevelConfig(ctx)
					if err != nil {
						log.CliLogger.Warn("Error in getting top level config")
						//If not found, set a default value
						config.CompatibilityLevel = "BACKWARD"
					}
				}
				mapOfMessageCompat["x-messageCompatibility"] = interface{}(config.CompatibilityLevel)
				//fmt.Println("Message Compatibility added")
				//Bindings
				bindings, opBindings, msgBindings, err := getBindings(cluster, topics[idx], clusterCreds, groupId)
				if err != nil {
					log.CliLogger.Warn("Bindings not found")

					return err
				}

				//Message
				(*spec.MessageEntity).WithBindings(entityProducer, spec.MessageBindingsObject{Kafka: &msgBindings})
				messages[strcase.ToCamel(topics[idx].Name)+"Message"] = spec.Message{OneOf1: &spec.MessageOneOf1{MessageEntity: (*spec.MessageEntity).WithPayload(entityProducer, producer)}}

				mustNotFail(reflector.AddChannel(asyncapi.ChannelInfo{
					Name: topics[idx].Name,
					BaseChannelItem: &spec.ChannelItem{
						MapOfAnything: mapOfMessageCompat,
						Subscribe: &spec.Operation{
							ID:      strcase.ToCamel(topics[idx].Name) + "Subscribe",
							Message: &spec.Message{Reference: &spec.Reference{Ref: "#/components/messages/" + strcase.ToCamel(topics[idx].Name) + "Message"}},
							//Message: &spec.Message{OneOf1: &spec.MessageOneOf1{MessageEntity: (*spec.MessageEntity).WithPayload(entityProducer, producer)}},
							Bindings: &spec.OperationBindingsObject{Kafka: &opBindings},
						},
						Bindings: &spec.ChannelBindingsObject{Kafka: &bindings},
					},
				}))

			}
		}
	}

	//Components
	reflector.Schema.WithComponents(spec.Components{Messages: messages,
		//Schemas: schemas,
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
	if getExamples {
		err = consumer.Close()
		if err != nil {
			log.CliLogger.Warn("Error in Closing the Consumer.")
			return err
		}
	}

	yaml, err := reflector.Schema.MarshalYAML()
	mustNotFail(err)
	fmt.Printf("\nSpec generated in file %s in current folder.\n", file)
	mustNotFail(ioutil.WriteFile(file, yaml, 0644))
	return nil
}

func getTags(schemaCluster *v1.SchemaRegistryCluster, prodSchema schemaregistry.Schema, apiKey, apiSecret string) ([]spec.Tag, error) {
	url := schemaCluster.SchemaRegistryEndpoint + "/catalog/v1/entity/type/sr_schema/name/" + schemaCluster.Id + ":.:" + strconv.Itoa(int(prodSchema.Id)) + "/tags"
	req, _ := http.NewRequest("GET", url, nil)
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
		log.CliLogger.Warn("Error in unmarshalling tags")
		log.CliLogger.Warn("Response received:", string(body))
		//log.CliLogger.
		return nil, err
	}
	var tags1 []spec.Tag

	for j := 0; j < len(tag); j++ {

		url = schemaCluster.SchemaRegistryEndpoint + "/catalog/v1/types/tagdefs/" + tag[j].TypeName
		req, _ = http.NewRequest("GET", url, nil)
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
		tags1 = append(tags1, spec.Tag{Name: tag[j].TypeName, Description: tagDef.Description})
	}
	return tags1, nil
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

		//fmt.Printf("Message value is \n%s\n", val)
		err = json.Unmarshal([]byte(val), &example)
		if err != nil {
			//logger.
			fmt.Printf("Example received for topic %s is not a valid JSON for unmarshalling.\n", topicName)
			log.CliLogger.Warn(err)
			return nil, err
		} else {
			//(*spec.MessageEntity).WithExamples(entityProducer, spec.MessageOneOf1OneOf1ExamplesItems{Payload: &example})
			return example, nil
		}
	}
}

func getBindings(cluster *schedv1.KafkaCluster, topic *schedv1.TopicDescription, clusterCreds *v1.APIKeyPair, groupId string) (interface{}, interface{}, interface{}, error) {
	var binding ConfluentBinding
	//Cleanup Policy
	var CleanupPolicy TopicConfigs
	var resp *http.Response
	url := cluster.RestEndpoint + "/kafka/v3/clusters/" + cluster.Id + "/topics/" + topic.Name + "/configs/cleanup.policy"
	req, _ := http.NewRequest("GET", url, nil)
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
		//fmt.Println(string(body))

		err := json.Unmarshal(body, &CleanupPolicy)
		if err != nil {
			fmt.Println("Error in Unmarshalling Topic Configs: Cleanup Policy")
			return nil, nil, nil, err
		}
	}

	//DeleteRetentionMs
	var DeleteRetentionMs TopicConfigs
	url = cluster.RestEndpoint + "/kafka/v3/clusters/" + cluster.Id + "/topics/" + topic.Name + "/configs/delete.retention.ms"
	req, _ = http.NewRequest("GET", url, nil)
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

func getClusterDetails(c *createCommand) (*schedv1.KafkaCluster, []*schedv1.TopicDescription, *v1.APIKeyPair, error) {
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
		return nil, nil, nil, err
	}
	clusterCreds := clusterConfig.APIKeys[clusterConfig.APIKey]
	if clusterCreds == nil {
		fmt.Println("Set an API Key Pair for the kafka Cluster using `confluent api-key create`")
		err = errors.New("API Key not set for the Kafka Cluster")
		return nil, nil, nil, err
	}
	topics, err := c.Client.Kafka.ListTopics(context.Background(), cluster)
	if err != nil {
		log.CliLogger.Warn("Error in getting topics")
		return nil, nil, nil, err
	}
	return cluster, topics, clusterCreds, nil
}
