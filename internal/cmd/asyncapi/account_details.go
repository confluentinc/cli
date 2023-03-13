package asyncapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	ckgo "github.com/confluentinc/confluent-kafka-go/kafka"
	schemaregistry "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/swaggest/go-asyncapi/spec-2.4.0"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type channelDetails struct {
	currentTopic            kafkarestv3.TopicData
	currentTopicDescription string
	currentSubject          string
	contentType             string
	schema                  *schemaregistry.Schema
	unmarshalledSchema      map[string]any
	mapOfMessageCompat      map[string]any
	topicLevelTags          []spec.Tag
	schemaLevelTags         []spec.Tag
	bindings                *bindings
	example                 any
}

type accountDetails struct {
	cluster        *ccstructs.KafkaCluster
	topics         []kafkarestv3.TopicData
	clusterCreds   *v1.APIKeyPair
	consumer       *ckgo.Consumer
	broker         string
	srCluster      *v1.SchemaRegistryCluster
	srClient       *schemaregistry.APIClient
	srContext      context.Context
	subjects       []string
	channelDetails channelDetails
}

const UserAgent = "User-Agent"

func (d *accountDetails) getTags() error {
	// Get topic level tags
	d.channelDetails.topicLevelTags = nil
	topicLevelTags, _, err := d.srClient.DefaultApi.GetTags(d.srContext, "kafka_topic", d.cluster.Id+":"+d.channelDetails.currentTopic.GetTopicName())
	if err != nil {
		return catchOpenAPIError(err)
	}
	for _, topicLevelTag := range topicLevelTags {
		d.channelDetails.topicLevelTags = append(d.channelDetails.topicLevelTags, spec.Tag{Name: topicLevelTag.TypeName})
	}

	// Get schema level tags
	d.channelDetails.schemaLevelTags = nil
	schemaLevelTags, _, err := d.srClient.DefaultApi.GetTags(d.srContext, "sr_schema", strconv.Itoa(int(d.channelDetails.schema.Id)))
	if err != nil {
		return catchOpenAPIError(err)
	}
	for _, schemaLevelTag := range schemaLevelTags {
		d.channelDetails.schemaLevelTags = append(d.channelDetails.schemaLevelTags, spec.Tag{Name: schemaLevelTag.TypeName})
	}
	return nil
}

func (d *accountDetails) getSchemaDetails() error {
	schema, _, err := d.srClient.DefaultApi.GetSchemaByVersion(d.srContext, d.channelDetails.currentSubject, "latest", nil)
	if err != nil {
		return err
	}
	d.channelDetails.schema = &schema
	if schema.SchemaType == "" {
		schema.SchemaType = "AVRO"
	}
	switch schema.SchemaType {
	case "JSON":
		d.channelDetails.contentType = "application/json"
	case "AVRO":
		d.channelDetails.contentType = "application/avro"
	case "PROTOBUF":
		return errors.New("protobuf is not supported")
	}
	// JSON or Avro Format
	err = json.Unmarshal([]byte(schema.Schema), &d.channelDetails.unmarshalledSchema)
	if err != nil {
		d.channelDetails.unmarshalledSchema, err = handlePrimitiveSchemas(schema.Schema, err)
		log.CliLogger.Warn(err)
	}
	return nil
}

func handlePrimitiveSchemas(schema string, err error) (map[string]any, error) {
	unmarshalledSchema := make(map[string]any)
	primitiveTypes := []string{"string", "null", "boolean", "int", "long", "float", "double", "bytes"}
	for _, primitiveType := range primitiveTypes {
		if schema == fmt.Sprintf("\"%s\"", primitiveType) {
			unmarshalledSchema["type"] = primitiveType
			return unmarshalledSchema, nil
		}
	}
	return nil, fmt.Errorf("failed to unmarshal schema: %v", err)
}

func (d *accountDetails) getTopicDescription() error {
	d.channelDetails.currentTopicDescription = ""
	atlasEntityWithExtInfo, _, err := d.srClient.DefaultApi.GetByUniqueAttributes(d.srContext, "kafka_topic", d.cluster.Id+":"+d.channelDetails.currentTopic.GetTopicName(), nil)
	if err != nil {
		return catchOpenAPIError(err)
	}
	if atlasEntityWithExtInfo.Entity.Attributes["description"] != nil {
		d.channelDetails.currentTopicDescription = fmt.Sprintf("%v", atlasEntityWithExtInfo.Entity.Attributes["description"])
	}
	return nil
}

func (c *command) countAsyncApiUsage(details *accountDetails) error {
	_, err := details.srClient.DefaultApi.AsyncapiPut(details.srContext)
	if err != nil {
		return fmt.Errorf("failed to access AsyncAPI metric endpoint: %v", err)
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
	if d.channelDetails.example != nil {
		entityProducer.WithExamples(spec.MessageOneOf1OneOf1ExamplesItems{Payload: &d.channelDetails.example})
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
	if openAPIError, ok := err.(schemaregistry.GenericOpenAPIError); ok {
		return errors.New(string(openAPIError.Body()))
	}
	return err
}
