package asyncapi

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/antihax/optional"
	"github.com/go-yaml/yaml"
	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	spec2 "github.com/swaggest/go-asyncapi/spec-2.4.0"
	yaml3 "k8s.io/apimachinery/pkg/util/yaml"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/confluentinc/cli/v3/pkg/retry"
	schemaregistry "github.com/confluentinc/cli/v3/pkg/schema-registry"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

const parseErrorMessage string = "topic is already present and `--overwrite` is not set"

type flagsImport struct {
	file                    string
	overwrite               bool
	kafkaApiKey             string
	schemaRegistryApiKey    string
	schemaRegistryApiSecret string
}

type kafkaBinding struct {
	BindingVersion     string                   `yaml:"bindingVersion"`
	Partitions         int32                    `yaml:"partitions"`
	TopicConfiguration topicConfigurationImport `yaml:"topicConfiguration"`
	XConfigs           map[string]string        `yaml:"x-configs"`
}

type topicConfigurationImport struct {
	CleanupPolicy       *[]string `yaml:"cleanup.policy"`
	RetentionTime       *int64    `yaml:"retention.ms"`
	RetentionSize       *int64    `yaml:"retention.bytes"`
	DeleteRetentionTime *int64    `yaml:"delete.retention.ms"`
	MaxMessageSize      *int32    `yaml:"max.message.bytes"`
}

type Message struct {
	SchemaFormat string      `yaml:"schemaFormat"`
	ContentType  string      `yaml:"contentType"`
	Payload      any         `yaml:"payload"`
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
	SecuritySchemes any                `yaml:"securitySchemes"`
}

type Security struct {
	ConfluentBroker         []any `yaml:"confluentBroker"`
	ConfluentSchemaRegistry []any `yaml:"confluentSchemaRegistry"`
}

type MessageRef struct {
	Ref string `yaml:"$ref"`
}

type Operation struct {
	OperationId string                        `yaml:"operationId"`
	TopicTags   []spec2.Tag                   `yaml:"tags"`
	OpBindings  spec2.OperationBindingsObject `yaml:"bindings"`
	Message     MessageRef                    `yaml:"message"`
}

type Topic struct {
	Description           string        `yaml:"description"`
	Publish               Operation     `yaml:"publish"`
	Subscribe             Operation     `yaml:"subscribe"`
	XMessageCompatibility string        `yaml:"x-messageCompatibility"`
	Bindings              TopicBindings `yaml:"bindings"`
}

type TopicBindings struct {
	Kafka kafkaBinding `yaml:"kafka"`
}

func (c *command) newImportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import an AsyncAPI specification.",
		Long:  "Update a Kafka cluster and Schema Registry according to an AsyncAPI specification file.",
		Args:  cobra.NoArgs,
		RunE:  c.asyncapiImport,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Import an AsyncAPI specification file to populate an existing Kafka cluster and Schema Registry.",
				Code: "confluent asyncapi import --file spec.yaml",
			},
		),
	}

	cmd.Flags().String("file", "", "Input filename.")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing topics with the same name.")
	cmd.Flags().String("kafka-api-key", "", "Kafka cluster API key.")

	// Deprecated
	cmd.Flags().String("schema-registry-api-key", "", "API key for Schema Registry.")
	cobra.CheckErr(cmd.Flags().MarkHidden("schema-registry-api-key"))

	// Deprecated
	cmd.Flags().String("schema-registry-api-secret", "", "API secret for Schema Registry.")
	cobra.CheckErr(cmd.Flags().MarkHidden("schema-registry-api-secret"))

	cobra.CheckErr(cmd.MarkFlagRequired("file"))
	cobra.CheckErr(cmd.MarkFlagFilename("file", "yaml", "yml"))

	return cmd
}

func getFlagsImport(cmd *cobra.Command) (*flagsImport, error) {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return nil, err
	}
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
		file:                    file,
		overwrite:               overwrite,
		kafkaApiKey:             kafkaApiKey,
		schemaRegistryApiKey:    schemaRegistryApiKey,
		schemaRegistryApiSecret: schemaRegistryApiSecret,
	}
	return flags, nil
}

func (c *command) asyncapiImport(cmd *cobra.Command, args []string) error {
	// Get flags
	flagsImp, err := getFlagsImport(cmd)
	if err != nil {
		return err
	}
	spec, err := fileToSpec(flagsImp.file)
	if err != nil {
		return err
	}
	details, err := c.getAccountDetails(cmd, &flags{kafkaApiKey: flagsImp.kafkaApiKey})
	if err != nil {
		return err
	}
	for topicName, topicDetails := range spec.Channels {
		if err := c.addChannelToCluster(details, spec, topicName, topicDetails.Bindings.Kafka, flagsImp.overwrite); err != nil {
			if err.Error() == parseErrorMessage {
				output.Printf("WARNING: topic \"%s\" is already present and `--overwrite` is not set.\n", topicName)
			} else {
				log.CliLogger.Warn(err)
			}
		}
	}
	return nil
}

func fileToSpec(fileName string) (*Spec, error) {
	asyncSpec, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	spec := new(Spec)
	if err := yaml.Unmarshal(asyncSpec, spec); err != nil {
		return nil, fmt.Errorf("unable to unmarshal YAML file: %v", err)
	}
	return spec, nil
}

func (c *command) addChannelToCluster(details *accountDetails, spec *Spec, topicName string, kafkaBinding kafkaBinding, overwrite bool) error {
	topicExistedAlready, newTopicCreated, err := c.addTopic(details, topicName, kafkaBinding, overwrite)
	if err != nil {
		log.CliLogger.Warn(err)
	}
	// If topic exists and overwrite flag is false, move to the next channel in spec
	if topicExistedAlready && !overwrite {
		return errors.New(parseErrorMessage)
	}
	// Register schema
	schemaId, err := registerSchema(details, topicName, spec.Components)
	if err != nil {
		return err
	}
	// Update subject compatibility
	if spec.Channels[topicName].XMessageCompatibility != "" {
		if err := updateSubjectCompatibility(details, spec.Channels[topicName].XMessageCompatibility, topicName+"-value"); err != nil {
			log.CliLogger.Warn(err)
		}
	}
	// Add tags
	if err := addSchemaTags(details, spec.Components, topicName, schemaId); err != nil {
		log.CliLogger.Warn(err)
	}
	if topicExistedAlready || newTopicCreated {
		if err := addTopicTags(details, spec.Channels[topicName].Subscribe, topicName); err != nil {
			log.CliLogger.Warn(err)
		}
	}
	// Add topic description to newly created topic
	if (topicExistedAlready || newTopicCreated) && spec.Channels[topicName].Description != "" {
		if err := addTopicDescription(details.srClient, fmt.Sprintf("%s:%s", details.kafkaClusterId, topicName),
			spec.Channels[topicName].Description); err != nil {
			return fmt.Errorf("unable to update topic description: %v", err)
		}
		output.Printf("Added description to topic \"%s\".\n", topicName)
	}
	return nil
}

func (c *command) addTopic(details *accountDetails, topicName string, kafkaBinding kafkaBinding, overwrite bool) (bool, bool, error) {
	// Check if topic already exists
	for _, topics := range details.topics {
		if topics.TopicName == topicName {
			// Topic already exists
			log.CliLogger.Warnf("Topic \"%s\" already exists.", topicName)
			if !overwrite {
				// Do not overwrite existing topic. Move to the next topic.
				return true, false, nil
			}
			// Overwrite existing topic
			err := c.updateTopic(topicName, kafkaBinding)
			return true, false, err
		}
	}
	// Create a new topic
	newTopicCreated, err := c.createTopic(topicName, kafkaBinding)
	return false, newTopicCreated, err
}

func (c *command) createTopic(topicName string, kafkaBinding kafkaBinding) (bool, error) {
	log.CliLogger.Infof("Topic not found. Adding a new topic: %s", topicName)
	topicConfigs := []kafkarestv3.CreateTopicRequestDataConfigs{}
	for configName, configValue := range combineTopicConfigs(kafkaBinding) {
		value := configValue
		topicConfigs = append(topicConfigs, kafkarestv3.CreateTopicRequestDataConfigs{
			Name:  configName,
			Value: *kafkarestv3.NewNullableString(&value),
		})
	}
	createTopicRequestData := kafkarestv3.CreateTopicRequestData{
		TopicName: topicName,
		Configs:   &topicConfigs,
	}
	if kafkaBinding.Partitions != 0 {
		createTopicRequestData.PartitionsCount = &kafkaBinding.Partitions
	}
	kafkaRest, err := c.GetKafkaREST()
	if err != nil {
		return false, err
	}
	if _, httpResp, err := kafkaRest.CloudClient.CreateKafkaTopic(createTopicRequestData); err != nil {
		restErr, parseErr := kafkarest.ParseOpenAPIErrorCloud(err)
		if parseErr == nil && restErr.Code == ccloudv2.BadRequestErrorCode {
			// Print partition limit error w/ suggestion
			if strings.Contains(restErr.Message, "partitions will exceed") {
				return false, errors.NewErrorWithSuggestions(restErr.Message, errors.ExceedPartitionLimitSuggestions)
			}
		}
		return false, kafkarest.NewError(kafkaRest.CloudClient.GetUrl(), err, httpResp)
	}
	output.Printf(errors.CreatedResourceMsg, resource.Topic, topicName)
	return true, nil
}

func (c *command) updateTopic(topicName string, kafkaBinding kafkaBinding) error {
	// Overwrite topic configs
	updateConfigs := []kafkarestv3.AlterConfigBatchRequestDataData{}
	modifiableConfigs := []string{}
	kafkaRest, err := c.GetKafkaREST()
	if err != nil {
		return err
	}
	configs, err := kafkaRest.CloudClient.ListKafkaTopicConfigs(topicName)
	if err != nil {
		return err
	}
	for _, configDetails := range configs.Data {
		if !configDetails.GetIsReadOnly() {
			modifiableConfigs = append(modifiableConfigs, configDetails.GetName())
		}
	}
	for configName, configValue := range combineTopicConfigs(kafkaBinding) {
		value := configValue
		if slices.Contains(modifiableConfigs, configName) {
			updateConfigs = append(updateConfigs, kafkarestv3.AlterConfigBatchRequestDataData{
				Name:  configName,
				Value: *kafkarestv3.NewNullableString(&value),
			})
		}
	}
	log.CliLogger.Info("Overwriting topic configs")
	if updateConfigs != nil {
		_, err = kafkaRest.CloudClient.UpdateKafkaTopicConfigBatch(topicName, kafkarestv3.AlterConfigBatchRequestData{Data: updateConfigs})
		if err != nil {
			return fmt.Errorf("unable to update topic configs: %v", err)
		}
	}
	output.Printf(errors.UpdatedResourceMsg, resource.Topic, topicName)
	return nil
}

// TopicConfiguration and XConfigs are both specifying Kafka configs, XConfigs is for those not specified in async api.
// This function combines TopicConfiguration and XConfigs
func combineTopicConfigs(kafkaBinding kafkaBinding) map[string]string {
	configs := make(map[string]string)
	if kafkaBinding.XConfigs != nil {
		configs = kafkaBinding.XConfigs
	}

	topicConfig := kafkaBinding.TopicConfiguration
	if topicConfig.CleanupPolicy != nil {
		configs["cleanup.policy"] = strings.Join(*topicConfig.CleanupPolicy, ",")
	}
	if topicConfig.RetentionTime != nil {
		configs["retention.ms"] = fmt.Sprint(*topicConfig.RetentionTime)
	}
	if topicConfig.RetentionSize != nil {
		configs["retention.bytes"] = fmt.Sprint(*topicConfig.RetentionSize)
	}
	if topicConfig.DeleteRetentionTime != nil {
		configs["delete.retention.ms"] = fmt.Sprint(*topicConfig.DeleteRetentionTime)
	}
	if topicConfig.MaxMessageSize != nil {
		configs["max.message.bytes"] = strconv.Itoa(int(*topicConfig.MaxMessageSize))
	}
	return configs
}

func addTopicDescription(client *schemaregistry.Client, qualifiedName, description string) error {
	atlasEntity := srsdk.AtlasEntityWithExtInfo{Entity: srsdk.AtlasEntity{
		Attributes: map[string]any{
			"description":   description,
			"qualifiedName": qualifiedName,
		},
		TypeName: "kafka_topic",
	}}
	return retry.Retry(5*time.Second, time.Minute, func() error {
		opts := &srsdk.PartialUpdateByUniqueAttributesOpts{AtlasEntityWithExtInfo: optional.NewInterface(atlasEntity)}
		return client.PartialUpdateByUniqueAttributes(opts)
	})
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
	if components.Messages != nil && components.Messages[strcase.ToCamel(topicName)+"Message"].Payload != nil {
		schema, err := yaml.Marshal(components.Messages[strcase.ToCamel(topicName)+"Message"].Payload)
		if err != nil {
			return 0, err
		}

		jsonSchema, err := yaml3.ToJSON(schema)
		if err != nil {
			return 0, fmt.Errorf("failed to encode schema as JSON: %v", err)
		}

		req := srsdk.RegisterSchemaRequest{
			Schema:     string(jsonSchema),
			SchemaType: resolveSchemaType(components.Messages[strcase.ToCamel(topicName)+"Message"].ContentType),
		}
		opts := &srsdk.RegisterOpts{Normalize: optional.NewBool(false)}
		id, err := details.srClient.Register(subject, req, opts)
		if err != nil {
			return 0, fmt.Errorf("unable to register schema: %v", err)
		}
		output.Printf("Registered schema \"%d\" under subject \"%s\".\n", id.Id, subject)
		return id.Id, nil
	}
	return 0, fmt.Errorf("schema payload not found in YAML input file")
}

func updateSubjectCompatibility(details *accountDetails, compatibility, subject string) error {
	// Updating the subject level compatibility
	log.CliLogger.Infof("Updating the Subject level compatibility to %s", compatibility)
	req := srsdk.ConfigUpdateRequest{Compatibility: compatibility}
	config, err := details.srClient.UpdateSubjectLevelConfig(subject, req)
	if err != nil {
		return fmt.Errorf("failed to update subject level compatibility: %v", err)
	}
	output.Printf("Subject level compatibility updated to \"%s\" for subject \"%s\".\n", config.Compatibility, subject)
	return nil
}

func addSchemaTags(details *accountDetails, components Components, topicName string, schemaId int32) error {
	// Schema level tags
	tagConfigs := []srsdk.Tag{}
	tagDefConfigs := []srsdk.TagDef{}
	tagNames := []string{}
	if components.Messages != nil {
		if components.Messages[strcase.ToCamel(topicName)+"Message"].Tags == nil {
			return nil
		}
		for _, tag := range components.Messages[strcase.ToCamel(topicName)+"Message"].Tags {
			tagDefConfigs = append(tagDefConfigs, srsdk.TagDef{
				// tag of type cf_entity so that it can be attached at any topic or schema level
				EntityTypes: []string{"cf_entity"},
				Name:        tag.Name,
			})
			tagConfigs = append(tagConfigs, srsdk.Tag{
				TypeName:   tag.Name,
				EntityType: "sr_schema",
				EntityName: fmt.Sprintf("%s:.:%s", details.schemaRegistryClusterId, strconv.Itoa(int(schemaId))),
			})
			tagNames = append(tagNames, tag.Name)
		}
		if err := addTagsUtil(details, tagDefConfigs, tagConfigs); err != nil {
			return err
		}
		output.Printf("Tag(s) %s added to schema \"%d\".\n", utils.ArrayToCommaDelimitedString(tagNames, "and"), schemaId)
	}
	return nil
}

func addTopicTags(details *accountDetails, subscribe Operation, topicName string) error {
	// Topic level tags
	// add topic level tags only if topic was created successfully or already exists
	if subscribe.TopicTags == nil {
		return nil
	}
	tagConfigs := []srsdk.Tag{}
	tagDefConfigs := []srsdk.TagDef{}
	tagNames := []string{}
	for _, tag := range subscribe.TopicTags {
		tagDefConfigs = append(tagDefConfigs, srsdk.TagDef{
			// tag of type cf_entity so that it can be attached at any topic or schema level
			EntityTypes: []string{"cf_entity"},
			Name:        tag.Name,
		})
		tagConfigs = append(tagConfigs, srsdk.Tag{
			TypeName:   tag.Name,
			EntityType: "kafka_topic",
			EntityName: fmt.Sprintf("%s:%s", details.kafkaClusterId, topicName),
		})
		tagNames = append(tagNames, tag.Name)
	}
	if err := addTagsUtil(details, tagDefConfigs, tagConfigs); err != nil {
		return err
	}
	output.Printf("Tag(s) %s added to Kafka topic \"%s\".\n", utils.ArrayToCommaDelimitedString(tagNames, "and"), topicName)
	return nil
}

func addTagsUtil(details *accountDetails, tagDefConfigs []srsdk.TagDef, tagConfigs []srsdk.Tag) error {
	tagDefOpts := &srsdk.CreateTagDefsOpts{TagDef: optional.NewInterface(tagDefConfigs)}
	err := retry.Retry(5*time.Second, time.Minute, func() error {
		_, err := details.srClient.CreateTagDefs(tagDefOpts)
		return err
	})
	if err != nil {
		return fmt.Errorf("unable to create tag definition: %v", err)
	}
	log.CliLogger.Debugf("Tag Definitions created")
	tagOpts := &srsdk.CreateTagsOpts{Tag: optional.NewInterface(tagConfigs)}
	err = retry.Retry(5*time.Second, time.Minute, func() error {
		_, err := details.srClient.CreateTags(tagOpts)
		return err
	})
	if err != nil {
		return fmt.Errorf("unable to add tag to resource: %v", err)
	}
	return nil
}
