package asyncapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	ckgo "github.com/confluentinc/confluent-kafka-go/kafka"
	schemaregistry "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/swaggest/go-asyncapi/spec-2.4.0"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type channelDetails struct {
	currentTopic       *schedv1.TopicDescription
	currentSubject     string
	contentType        string
	schema             *schemaregistry.Schema
	unmarshalledSchema map[string]interface{}
	mapOfMessageCompat map[string]interface{}
	topicLevelTags     []spec.Tag
	schemaLevelTags    []spec.Tag
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

func (d *accountDetails) getTags() error {
	topicLevelTags, _, err := d.srClient.DefaultApi.GetTags(d.srContext, "kafka_topic", d.cluster.Id+":"+d.channelDetails.currentTopic.Name)
	if err != nil {
		return fmt.Errorf("failed to get topic level tags: %v", err)
	}
	var topicLevelTagsInSpec []spec.Tag
	for _, topicLevelTag := range topicLevelTags {
		topicLevelTagsInSpec = append(topicLevelTagsInSpec, spec.Tag{Name: topicLevelTag.TypeName})
	}
	d.channelDetails.topicLevelTags = topicLevelTagsInSpec
	// Get Schema level tags
	schemaLevelTags, _, err := d.srClient.DefaultApi.GetTags(d.srContext, "sr_schema", strconv.Itoa(int(d.channelDetails.schema.Id)))
	if err != nil {
		return fmt.Errorf("failed to get schema level tags: %v", err)
	}
	var schemaLevelTagsInSpec []spec.Tag
	for _, schemaLevelTag := range schemaLevelTags {
		schemaLevelTagsInSpec = append(schemaLevelTagsInSpec, spec.Tag{Name: schemaLevelTag.TypeName})
	}
	d.channelDetails.schemaLevelTags = schemaLevelTagsInSpec
	return nil
}

func (d *accountDetails) getSchemaDetails() error {
	log.CliLogger.Debugf("Adding operation: %s", d.channelDetails.currentTopic.Name)
	schema, _, err := d.srClient.DefaultApi.GetSchemaByVersion(d.srContext, d.channelDetails.currentSubject, "latest", nil)
	if err != nil {
		return err
	}
	var unmarshalledSchema map[string]interface{}
	if schema.SchemaType == "" {
		d.channelDetails.contentType = "application/avro"
	} else if schema.SchemaType == "JSON" {
		d.channelDetails.contentType = "application/json"
	} else if schema.SchemaType == "PROTOBUF" {
		log.CliLogger.Warn("Protobuf not supported.")
		d.channelDetails.contentType = "PROTOBUF"
		return nil
	}
	// JSON or Avro Format
	err = json.Unmarshal([]byte(schema.Schema), &unmarshalledSchema)
	if err != nil {
		return fmt.Errorf("failed to unmarshal schema: %v", err)

	}
	d.channelDetails.unmarshalledSchema = unmarshalledSchema
	d.channelDetails.schema = &schema
	return nil
}

func (d *accountDetails) buildMessageEntity() *spec.MessageEntity {
	entityProducer := new(spec.MessageEntity)
	(*spec.MessageEntity).WithContentType(entityProducer, d.channelDetails.contentType)
	if d.channelDetails.contentType == "application/avro" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/vnd.apache.avro;version=1.9.0")
	} else if d.channelDetails.contentType == "application/json" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/schema+json;version=draft-07")
	}
	(*spec.MessageEntity).WithTags(entityProducer, d.channelDetails.schemaLevelTags...)
	// Name
	(*spec.MessageEntity).WithName(entityProducer, msgName(d.channelDetails.currentTopic.Name))
	// Example
	if d.channelDetails.example != nil {
		(*spec.MessageEntity).WithExamples(entityProducer, spec.MessageOneOf1OneOf1ExamplesItems{Payload: &d.channelDetails.example})
	}
	(*spec.MessageEntity).WithBindings(entityProducer, d.channelDetails.bindings.messageBinding)
	(*spec.MessageEntity).WithPayload(entityProducer, d.channelDetails.unmarshalledSchema)
	return entityProducer
}
