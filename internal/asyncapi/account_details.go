package asyncapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/swaggest/go-asyncapi/spec-2.4.0"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/schemaregistry"
)

type channelDetails struct {
	currentTopic            kafkarestv3.TopicData
	currentTopicDescription string
	currentSubject          string
	contentType             string
	schema                  *srsdk.Schema
	unmarshalledSchema      map[string]any
	mapOfMessageCompat      map[string]any
	topicLevelTags          []spec.Tag
	schemaLevelTags         []spec.Tag
	bindings                *bindings
	examples                map[string]any
}

type accountDetails struct {
	kafkaClusterId          string
	schemaRegistryClusterId string
	topics                  []kafkarestv3.TopicData
	clusterCreds            *config.APIKeyPair
	consumer                *ckgo.Consumer
	kafkaUrl                string
	kafkaBootstrapUrl       string
	schemaRegistryUrl       string
	srClient                *schemaregistry.Client
	subjects                []string
	channelDetails          channelDetails
}

func (d *accountDetails) getTags() error {
	// Get topic level tags
	d.channelDetails.topicLevelTags = nil
	topicLevelTags, err := d.srClient.GetTags("kafka_topic", d.kafkaClusterId+":"+d.channelDetails.currentTopic.GetTopicName())
	if err != nil {
		return catchOpenAPIError(err)
	}
	for _, topicLevelTag := range topicLevelTags {
		d.channelDetails.topicLevelTags = append(d.channelDetails.topicLevelTags, spec.Tag{Name: topicLevelTag.GetTypeName()})
	}

	// Get schema level tags
	d.channelDetails.schemaLevelTags = nil
	schemaLevelTags, err := d.srClient.GetTags("sr_schema", strconv.Itoa(int(d.channelDetails.schema.GetId())))
	if err != nil {
		return catchOpenAPIError(err)
	}
	for _, schemaLevelTag := range schemaLevelTags {
		d.channelDetails.schemaLevelTags = append(d.channelDetails.schemaLevelTags, spec.Tag{Name: schemaLevelTag.GetTypeName()})
	}
	return nil
}

func (d *accountDetails) getSchemaDetails() error {
	schema, err := d.srClient.GetSchemaByVersion(d.channelDetails.currentSubject, "latest", false)
	if err != nil {
		return err
	}
	d.channelDetails.schema = &schema

	if schema.GetSchemaType() == "" {
		schema.SchemaType = srsdk.PtrString("AVRO")
	}

	// The backend considers "AVRO" to be the default schema type.
	switch schema.GetSchemaType() {
	case "PROTOBUF":
		return fmt.Errorf("protobuf is not supported")
	case "AVRO", "JSON":
		d.channelDetails.contentType = fmt.Sprintf("application/%s", strings.ToLower(schema.GetSchemaType()))
	}

	if err := json.Unmarshal([]byte(schema.GetSchema()), &d.channelDetails.unmarshalledSchema); err != nil {
		d.channelDetails.unmarshalledSchema, err = handlePrimitiveSchemas(schema.GetSchema(), err)
		log.CliLogger.Warn(err)
	}

	return nil
}

func handlePrimitiveSchemas(schema string, err error) (map[string]any, error) {
	unmarshalledSchema := make(map[string]any)
	primitiveTypes := []string{"string", "null", "boolean", "int", "long", "float", "double", "bytes"}
	for _, primitiveType := range primitiveTypes {
		if schema == fmt.Sprintf(`"%s"`, primitiveType) {
			unmarshalledSchema["type"] = primitiveType
			return unmarshalledSchema, nil
		}
	}
	return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
}

func (d *accountDetails) getTopicDescription() error {
	d.channelDetails.currentTopicDescription = ""
	atlasEntityWithExtInfo, err := d.srClient.GetByUniqueAttributes("kafka_topic", d.kafkaClusterId+":"+d.channelDetails.currentTopic.GetTopicName())
	if err != nil {
		return catchOpenAPIError(err)
	}
	attributes := atlasEntityWithExtInfo.Entity.GetAttributes()
	if _, ok := attributes["description"]; ok {
		d.channelDetails.currentTopicDescription = fmt.Sprintf("%v", attributes["description"])
	}
	return nil
}

func (d *accountDetails) buildMessageEntity() *spec.MessageEntity {
	entityProducer := new(spec.MessageEntity)
	entityProducer.WithContentType(d.channelDetails.contentType)
	if d.channelDetails.contentType == "application/avro" {
		entityProducer.WithSchemaFormat("application/vnd.apache.avro;version=1.9.0")
	} else if d.channelDetails.contentType == "application/json" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/schema+json;version=draft-07")
	}
	entityProducer.WithTags(d.channelDetails.schemaLevelTags...)
	// Name
	entityProducer.WithName(msgName(d.channelDetails.currentTopic.GetTopicName()))
	// Example
	if d.channelDetails.examples != nil && d.channelDetails.examples[d.channelDetails.currentTopic.GetTopicName()] != nil {
		topicExample := d.channelDetails.examples[d.channelDetails.currentTopic.GetTopicName()]
		entityProducer.WithExamples(spec.MessageOneOf1OneOf1ExamplesItems{Payload: &topicExample})
	}
	if d.channelDetails.bindings != nil {
		entityProducer.WithBindings(d.channelDetails.bindings.messageBinding)
	}
	if d.channelDetails.unmarshalledSchema != nil {
		entityProducer.WithPayload(d.channelDetails.unmarshalledSchema)
	}
	return entityProducer
}

func catchOpenAPIError(err error) error {
	if openAPIError, ok := err.(srsdk.GenericOpenAPIError); ok {
		return errors.New(string(openAPIError.Body()))
	}
	return err
}
