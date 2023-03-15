package asyncapi

import (
	"context"
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
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

const parseErrorMessage string = "topic is already present and overwrite flag is false. Moving to the next topic"

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
	XConfigs map[string]string `yaml:"x-configs"`
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
	ConfluentBroker         []any `yaml:"confluentBroker"`
	ConfluentSchemaRegistry []any `yaml:"confluentSchemaRegistry"`
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
	Description           string        `yaml:"description"`
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
		Use:   "import <spec.yaml>",
		Short: "Adds topics, schemas and tags from the yaml file.",
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{Text: "Import an asyncapi specification file to populate an existing cluster",
				Code: "confluent asyncapi import spec.yaml",
			},
		),
	}
	c := &command{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner), nil}
	cmd.RunE = c.parse
	cmd.Flags().Bool("overwrite", false, "Overwrite existing topics and schemas.")
	cmd.Flags().String("kafka-api-key", "", "Kafka cluster API key.")
	cmd.Flags().String("schema-registry-api-key", "", "API key for Schema Registry.")
	cmd.Flags().String("schema-registry-api-secret", "", "API secret for Schema Registry.")

	return cmd
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

	flags := &flagsImport{
		overwrite:               overwrite,
		kafkaApiKey:             kafkaApiKey,
		schemaRegistryApiKey:    schemaRegistryApiKey,
		schemaRegistryApiSecret: schemaRegistryApiSecret,
	}
	return flags, nil
}

func (c *command) parse(cmd *cobra.Command, args []string) error {
	spec, err := fileToStruct(args[0])
	if err != nil {
		return err
	}
	// Get flags
	flagsImp, err := getFlagsImport(cmd)
	if err != nil {
		return err
	}
	// Getting Kafka Cluster & SR
	flags := createExportFlagsFromImportFlags(flagsImp)
	details, err := c.getAccountDetails(&flags)
	if err != nil {
		return err
	}
	for topicName, topicDetails := range spec.Channels {
		err = c.addChannelToCluster(details, spec, topicName, topicDetails.TopicBinding.Kafka, false, flagsImp.overwrite)
		if err != nil {
			if err.Error() == parseErrorMessage {
				log.CliLogger.Info(err)
			} else {
				log.CliLogger.Warn(err)
			}
		}
	}
	output.Printf("AsyncAPI specification imported from \"%s\"\n", args[0])
	return nil
}

func fileToStruct(fileName string) (*Spec, error) {
	asyncSpec, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	spec := new(Spec)
	err = yaml.Unmarshal(asyncSpec, &spec)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal yaml file: %v", err)
	}
	return spec, nil
}

func (c *command) addChannelToCluster(details *accountDetails, spec *Spec, topicName string, kafkaBinding KafkaBinding, topicNotCreated, overwrite bool) error {
	output.Printf("Importing topic: %s\n", topicName)
	topicExists, err := c.addTopic(details, topicName, spec.Channels[topicName].Description, kafkaBinding, overwrite)
	if err != nil {
		// unable to create kafka topic
		log.CliLogger.Warn(err)
		topicNotCreated = true
	}
	// If topic exists and overwrite flag is false, move to the next channel in spec
	if topicExists && !overwrite {
		return errors.New(parseErrorMessage)
	}
	// Register schema
	schemaId, err := registerSchema(details, topicName, spec.Components)
	if err != nil {
		return err
	}
	// Update subject compatibility
	if err := updateSubjectCompatibility(details, spec.Channels[topicName].XMessageCompatibility, topicName+"-value"); err != nil {
		return err
	}
	// Add tags
	output.Println("Adding tags at schema level and topic level.")
	if _, _, err := addSchemaTags(details, spec.Components, topicName, schemaId); err != nil {
		return err
	}
	if !topicNotCreated && spec.Channels != nil {
		if _, _, err := addTopicTags(details, spec.Channels[topicName].Subscribe, topicName); err != nil {
			return err
		}
	}
	return nil
}

func createExportFlagsFromImportFlags(flagsImp *flagsImport) flags {
	flagsExport := flags{
		kafkaApiKey:             flagsImp.kafkaApiKey,
		schemaRegistryApiKey:    flagsImp.schemaRegistryApiKey,
		schemaRegistryApiSecret: flagsImp.schemaRegistryApiSecret,
	}
	return flagsExport
}

func isModifiableConfig(configName string) bool {
	modifiableConfigs := []string{"delete.retention.ms", "max.compaction.lag.ms", "max.message.bytes", "message.timestamp.difference.max.ms", "message.timestamp.type",
		"min.compaction.lag.ms", "min.insync.replicas", "retention.bytes", "retention.ms", "segment.bytes", "segment.ms"}
	for _, config := range modifiableConfigs {
		if config == configName {
			return true
		}
	}
	return false
}

func (c *command) addTopic(details *accountDetails, topicName string, description string, kb KafkaBinding, overwrite bool) (bool, error) {
	topicConfigs := []kafkarestv3.CreateTopicRequestDataConfigs{}
	updateConfigs := []kafkarestv3.AlterConfigBatchRequestDataData{}
	for configName, configValue := range kb.XConfigs {
		value := configValue
		topicConfigs = append(topicConfigs, kafkarestv3.CreateTopicRequestDataConfigs{
			Name:  configName,
			Value: *kafkarestv3.NewNullableString(&value),
		})
		if isModifiableConfig(configName) {
			updateConfigs = append(updateConfigs, kafkarestv3.AlterConfigBatchRequestDataData{
				Name:  configName,
				Value: *kafkarestv3.NewNullableString(&value),
			})
		}
	}
	topicRequestData := kafkarestv3.CreateTopicRequestData{
		TopicName: topicName,
		Configs:   &topicConfigs,
	}
	// Check if topic already exists
	for _, topics := range details.topics {
		if topics.TopicName == topicName {
			// Topic already exists
			log.CliLogger.Info("Topic is already present.")
			if !overwrite {
				// Do not overwrite existing topic. Move to the next topic.
				return true, nil
			}
			// Overwrite topic configs
			log.CliLogger.Info("Overwriting topic configs")
			_, err := c.kafkaRest.CloudClient.UpdateKafkaTopicConfigBatch(details.clusterId, topicName, kafkarestv3.AlterConfigBatchRequestData{Data: updateConfigs})
			if err != nil {
				return true, fmt.Errorf("unable to update topic config delete.retention.ms: %v", err)
			}
			if description != "" {
				err = addTopicDescription(details.srClient, details.srContext, details.clusterId+":"+topicName, description)
			}
			return true, err
		}
	}
	log.CliLogger.Infof("Topic not found. Adding a new topic: %s", topicName)
	_, httpResp, err := c.kafkaRest.CloudClient.CreateKafkaTopic(details.clusterId, topicRequestData)
	if err != nil {
		restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
		if parseErr == nil && restErr.Code == ccloudv2.BadRequestErrorCode {
			// Print partition limit error w/ suggestion
			if strings.Contains(restErr.Message, "partitions will exceed") {
				return false, errors.NewErrorWithSuggestions(restErr.Message, errors.ExceedPartitionLimitSuggestions)
			}
		}
		return false, kafkarest.NewError(c.kafkaRest.CloudClient.GetUrl(), err, httpResp)
	}
	log.CliLogger.Infof("Added topic %s to cluster %s.\n", topicName, details.clusterId)

	if description != "" {
		err = addTopicDescription(details.srClient, details.srContext, details.clusterId+":"+topicName, description)
	}
	return false, err
}

func addTopicDescription(srClient *srsdk.APIClient, srContext context.Context, qualifiedName string, description string) error {
	atlasEntity := srsdk.AtlasEntityWithExtInfo{
		ReferredEntities: nil,
		Entity: srsdk.AtlasEntity{
			Attributes: map[string]interface{}{"description": description, "qualifiedName": qualifiedName},
			TypeName:   "kafka_topic",
		},
	}
	_, err := srClient.DefaultApi.PartialUpdateByUniqueAttributes(srContext,
		&srsdk.PartialUpdateByUniqueAttributesOpts{AtlasEntityWithExtInfo: optional.NewInterface(atlasEntity)})
	return err
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
	// Registering the schema
	// Subject name follows the TopicNameStrategy
	subject := topicName + "-value"
	output.Printf("Registering schema under the subject %s\n", subject)
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
	// Updating the subject level compatibility
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
	// Schema level tags
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
	// Topic level tags
	// add topic level tags only if topic was created successfully or already exists
	tagConfigs := new([]srsdk.Tag)
	tagDefConfigs := new([]srsdk.TagDef)
	for _, tag := range subscribe.TopicTags {
		*tagDefConfigs = append(*tagDefConfigs, srsdk.TagDef{
			EntityTypes: []string{"cf_entity"},
			Name:        tag.Name,
		})
		*tagConfigs = append(*tagConfigs, srsdk.Tag{
			TypeName:   tag.Name,
			EntityType: "kafka_topic",
			EntityName: details.srCluster.Id + ":" + details.clusterId + ":" + topicName,
		})
	}
	err := addTagsUtil(details, *tagDefConfigs, *tagConfigs)
	return *tagConfigs, *tagDefConfigs, err
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
