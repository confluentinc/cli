package asyncapi

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/antihax/optional"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/go-yaml/yaml"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	spec2 "github.com/swaggest/go-asyncapi/spec-2.4.0"
	yaml3 "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type flagsImport struct {
	overwrite               bool
	kafkaApiKey             string
	schemaRegistryApiKey    string
	schemaRegistryApiSecret string
}

type BindingsXConfigs struct {
	CleanupPolicy                  string `yaml:"cleanup.policy"`
	DeleteRetentionMs              int    `yaml:"delete.retention.ms"`
	ConfluentValueSchemaValidation string `yaml:"confluent.value.schema.validation"`
}

type KafkaBinding struct {
	XPartitions int              `yaml:"x-partitions"`
	XReplicas   int              `yaml:"x-replicas"`
	XConfigs    BindingsXConfigs `yaml:"x-configs"`
}

type Message struct {
	SchemaFormat string      `yaml:"schemaFormat"`
	ContentType  string      `yaml:"contentType"`
	Payload      interface{} `yaml:"payload"`
	Name         string      `yaml:"name"`
	Tags         []spec2.Tag `yaml:"tags"`
}

type Spec struct {
	Asyncapi   string           `yaml:"asyncapi"`
	Info       spec2.Info       `yaml:"info"`
	Servers    spec2.Server     `yaml:"servers"`
	Channels   map[string]Topic `yaml:"channels"`
	Components Components       `yaml:"components"`
}

type Components struct {
	Messages        map[string]Message `yaml:"messages"`
	SecuritySchemes interface{}        `yaml:"securitySchemes"`
}

type Security struct {
	ConfluentBroker         []interface{} `yaml:"confluentBroker"`
	ConfluentSchemaRegistry []interface{} `yaml:"confluentSchemaRegistry"`
}

type MessageRef struct {
	Ref string `yaml:"$ref"`
}
type Operation struct {
	OperationID string                        `yaml:"operationId"`
	TopicTags   []spec2.Tag                   `yaml:"tags"`
	OpBindings  spec2.OperationBindingsObject `yaml:"bindings"`
	Message     MessageRef                    `yaml:"message"`
}

type Topic struct {
	Publish               Operation     `yaml:"publish"`
	Subscribe             Operation     `yaml:"subscribe"`
	XMessageCompatibility string        `yaml:"x-messageCompatibility"`
	TopicBinding          TopicBindings `yaml:"bindings"`
}

type TopicBindings struct {
	Kafka KafkaBinding `yaml:"kafka"`
}

type ConfluentBrokerXConfigs struct {
	SaslMechanisms   string `yaml:"sasl.mechanisms"`
	SaslPassword     string `yaml:"sasl.password"`
	SaslUsername     string `yaml:"sasl.username"`
	SecurityProtocol string `yaml:"security.protocol"`
}
type ConfluentBroker struct {
	Type     string                  `yaml:"type"`
	XConfigs ConfluentBrokerXConfigs `yaml:"x-configs"`
}

type ConfluentSRXConfigs struct {
	BasicAuthUserInfo string `yaml:"basic.auth.user.info"`
}
type ConfluentSchemaRegistry struct {
	Type     string              `yaml:"type"`
	XConfigs ConfluentSRXConfigs `yaml:"x-configs"`
}

func newImportCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "import spec.yaml",
		Short:   "Adds topics, schemas and tags from the yaml file.",
		Args:    cobra.ExactArgs(1),
		Example: "confluent asyncapi import spec.yaml",
	}
	c := &command{AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}
	c.RunE = c.reverse
	c.Flags().Bool("overwrite", false, "overwrite existing topics and schemas")
	c.Flags().String("kafka-api-key", "", "Kafka cluster API key.")
	c.Flags().String("schema-registry-api-key", "", "API key for Schema Registry.")
	c.Flags().String("schema-registry-api-secret", "", "API secret for Schema Registry.")
	return c.Command
}

func getFlagsImport(cmd *cobra.Command) (*flagsImport, error) {
	overwrite, err := cmd.Flags().GetBool("overwrite")
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

	return &flagsImport{
		overwrite:               overwrite,
		kafkaApiKey:             kafkaApiKey,
		schemaRegistryApiKey:    schemaRegistryApiKey,
		schemaRegistryApiSecret: schemaRegistryApiSecret,
	}, nil
}

func (c *command) reverse(cmd *cobra.Command, strings []string) error {
	fileName := os.Args[3]
	asyncSpec, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	spec := new(Spec)
	err = yaml.Unmarshal(asyncSpec, &spec)
	if err != nil {
		return fmt.Errorf("unable to unmarshal yaml file: %v", err)
	}
	// Get flags
	flagsImp, err := getFlagsImport(cmd)
	if err != nil {
		return err
	}
	//Getting Kafka Cluster & SR
	flags := new(flags)
	flags.kafkaApiKey = flagsImp.kafkaApiKey
	flags.schemaRegistryApiKey = flagsImp.schemaRegistryApiKey
	flags.schemaRegistryApiSecret = flagsImp.schemaRegistryApiSecret
	details, err := c.getAccountDetails(flags)
	if err != nil {
		return err
	}
	var topicExists = false
	var topicNotCreated = false
	//Add Topics
	for topicName, topicDetails := range spec.Channels {
		//For each topic in spec
		//Reset topicExists variable
		topicExists = false
		topicNotCreated = false
		output.Printf("Importing topic: %s", topicName)
		topicExists, err = addTopic(details, topicName, topicDetails.TopicBinding.Kafka, flagsImp.overwrite)
		if err != nil {
			//unable to create kafka topic
			log.CliLogger.Debugf("unable to create kafka topic: %v", err)
			topicNotCreated = true
		}
		//If topic exists and overwrite flag is false, move to the next channel in spec
		if topicExists && !flagsImp.overwrite {
			continue
		}
		//Register schema
		schemaId, err := registerSchema(details, topicName, spec.Components)
		if err != nil {
			log.CliLogger.Warn(err)
		}
		//Update subject compatibility
		if err := updateSubjectCompatibility(details, spec.Channels[topicName].XMessageCompatibility, topicName+"-value"); err != nil {
			log.CliLogger.Warn(err)
		}
		//Add tags
		if _, _, err := addSchemaTags(details, spec.Components, topicName, schemaId); err != nil {
			log.CliLogger.Warn(err)
		}
		if !topicNotCreated && spec.Channels != nil {
			if _, _, err := addTopicTags(details, spec.Channels[topicName].Subscribe, topicName); err != nil {
				log.CliLogger.Warn(err)
			}
		}
	}
	return nil
}

func addTopic(details *accountDetails, topicName string, kb KafkaBinding, overwrite bool) (bool, error) {
	deleteRetentionMs := strconv.FormatInt(int64(kb.XConfigs.DeleteRetentionMs), 10)
	cleanupPolicy := kb.XConfigs.CleanupPolicy
	schemaValidation := kb.XConfigs.ConfluentValueSchemaValidation
	topicConfigs := []kafkarestv3.CreateTopicRequestDataConfigs{
		{Name: "delete.retention.ms", Value: *kafkarestv3.NewNullableString(&deleteRetentionMs)},
		{Name: "cleanup.policy", Value: *kafkarestv3.NewNullableString(&cleanupPolicy)},
		{Name: "confluent.key.schema.validation", Value: *kafkarestv3.NewNullableString(&schemaValidation)},
	}
	updateConfigs := []kafkarestv3.AlterConfigBatchRequestDataData{
		{Name: "delete.retention.ms",
			Value: *kafkarestv3.NewNullableString(&deleteRetentionMs)},
	}

	topicRequestData := kafkarestv3.CreateTopicRequestData{
		TopicName:         topicName,
		Configs:           &topicConfigs,
		PartitionsCount:   utils.Int32Ptr(int32(kb.XPartitions)),
		ReplicationFactor: utils.Int32Ptr(int32(kb.XReplicas)),
	}
	//Check if topic already exists
	for _, topics := range details.topics {
		if topics.TopicName == topicName {
			//Topic already exists
			log.CliLogger.Info("Topic is already present.")
			if overwrite == false {
				// Do not overwrite existing topic. Move to the next topic.
				return true, nil
			}
			// Overwrite topic configs
			log.CliLogger.Info("Overwriting topic config delete.retention.ms with value in spec")
			_, err := details.kafkaRest.CloudClient.UpdateKafkaTopicConfigBatch(details.cluster.Id, topicName, kafkarestv3.AlterConfigBatchRequestData{Data: updateConfigs})
			if err != nil {
				return true, fmt.Errorf("unable to update topic config delete.retention.ms: %v", err)
			}
		}
	}
	log.CliLogger.Infof("Topic not found. Adding a new topic: %s", topicName)
	_, httpResp, err := details.kafkaRest.CloudClient.CreateKafkaTopic(details.cluster.Id, topicRequestData)
	if err != nil {
		restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
		if parseErr == nil && restErr.Code == ccloudv2.BadRequestErrorCode {
			// Print partition limit error w/ suggestion
			if strings.Contains(restErr.Message, "partitions will exceed") {
				return false, errors.NewErrorWithSuggestions(restErr.Message, errors.ExceedPartitionLimitSuggestions)
			}
		}
		return false, kafkarest.NewError(details.kafkaRest.CloudClient.GetUrl(), err, httpResp)
	}
	log.CliLogger.Infof("Added topic %s to cluster %s.\n", topicName, details.cluster.Id)
	return false, nil
}

func resolveSchemaType(contentType string) string {
	if strings.Contains(contentType, "avro") {
		return "AVRO"
	} else if strings.Contains(contentType, "json") {
		return "JSON"
	} else {
		return "PROTOBUF"
	}
}

func registerSchema(details *accountDetails, topicName string, components Components) (int32, error) {
	//Registering the schema
	//Subject name follows the TopicNameStrategy
	subject := topicName + "-value"
	if components.Messages != nil && components.Messages[strcase.ToCamel(topicName)+"Message"].Payload != nil {
		schema, err := yaml.Marshal(components.Messages[strcase.ToCamel(topicName)+"Message"].Payload)
		if err != nil {
			return 0, err
		}
		jsonSchema, err := yaml3.ToJSON(schema)
		if err != nil {
			return 0, fmt.Errorf("error in converting schema to JSON format: %v", err)
		}
		id, _, err := details.srClient.DefaultApi.Register(details.srContext, subject,
			srsdk.RegisterSchemaRequest{
				Schema:     string(jsonSchema),
				SchemaType: resolveSchemaType(components.Messages[strcase.ToCamel(topicName)+"Message"].ContentType),
			})
		if err != nil {
			return 0, fmt.Errorf("unable to register schema due to %v", err)
		}
		log.CliLogger.Infof("Schema registered under subject %s with ID %d.\n", subject, id.Id)
		return id.Id, nil
	}
	return 0, fmt.Errorf("schema payload not found in yaml input file")
}

func updateSubjectCompatibility(details *accountDetails, compatibility string, subject string) error {
	//Updating the subject level compatibility
	log.CliLogger.Infof("updating the Subject level compatibility to %s.\n", compatibility)
	updateReq := srsdk.ConfigUpdateRequest{Compatibility: compatibility}
	config, _, err := details.srClient.DefaultApi.UpdateSubjectLevelConfig(details.srContext, subject, updateReq)
	if err != nil {
		return fmt.Errorf("failed to update subject level compatibility: %v", err)
	}
	log.CliLogger.Infof("subject level compatibility updated to %s", config.Compatibility)
	return nil
}

func addSchemaTags(details *accountDetails, components Components, topicName string, schemaId int32) ([]srsdk.Tag, []srsdk.TagDef, error) {
	//Schema level tags
	var tagConfigs []srsdk.Tag
	var tagDefConfigs []srsdk.TagDef
	if components.Messages != nil {
		for _, tag := range components.Messages[strcase.ToCamel(topicName)+"Message"].Tags {
			tagDefConfigs = append(tagDefConfigs, srsdk.TagDef{
				EntityTypes: []string{"cf_entity"},
				Name:        tag.Name,
			})
			tagConfigs = append(tagConfigs, srsdk.Tag{
				TypeName:   tag.Name,
				EntityType: "sr_schema",
				EntityName: details.srCluster.Id + ":.:" + strconv.Itoa(int(schemaId)),
			})
		}
		err := addTagsUtil(details, tagDefConfigs, tagConfigs)
		return tagConfigs, tagDefConfigs, err
	}
	return tagConfigs, tagDefConfigs, nil
}

func addTopicTags(details *accountDetails, subscribe Operation, topicName string) ([]srsdk.Tag, []srsdk.TagDef, error) {
	//Topic level tags
	//add topic level tags only if topic was created successfully or already exists
	var tagConfigs []srsdk.Tag
	var tagDefConfigs []srsdk.TagDef
	for _, tag := range subscribe.TopicTags {
		tagDefConfigs = append(tagDefConfigs, srsdk.TagDef{
			EntityTypes: []string{"cf_entity"},
			Name:        tag.Name,
		})
		tagConfigs = append(tagConfigs, srsdk.Tag{
			TypeName:   tag.Name,
			EntityType: "kafka_topic",
			EntityName: details.srCluster.Id + ":" + details.cluster.Id + ":" + topicName,
		})
	}
	err := addTagsUtil(details, tagDefConfigs, tagConfigs)
	return tagConfigs, tagDefConfigs, err
}

func addTagsUtil(details *accountDetails, tagDefConfigs []srsdk.TagDef, tagConfigs []srsdk.Tag) error {
	tagDefOpts := new(srsdk.CreateTagDefsOpts)
	tagDefOpts.TagDef = optional.NewInterface(tagDefConfigs)
	defs, _, err := details.srClient.DefaultApi.CreateTagDefs(details.srContext, tagDefOpts)
	if err != nil {
		return fmt.Errorf("unable to create tag definition: %v", err)
	}
	log.CliLogger.Debugf("Tag Definitions created: %v", defs)
	tagOpts := new(srsdk.CreateTagsOpts)
	tagOpts.Tag = optional.NewInterface(tagConfigs)
	tags, _, err := details.srClient.DefaultApi.CreateTags(details.srContext, tagOpts)
	log.CliLogger.Debugf("%v added to resource %s", tags, tagConfigs[0].EntityName)
	if err != nil {
		return fmt.Errorf("unable to add tag to resource: %v", err)
	}
	log.CliLogger.Infof("%v added to resource %s", tags, tagConfigs[0].EntityName)
	return nil
}
