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
	currentTopic            *schedv1.TopicDescription
	currentTopicDescription interface{}
	currentSubject          string
	contentType             string
	schema                  *schemaregistry.Schema
	unmarshalledSchema      map[string]interface{}
	mapOfMessageCompat      map[string]interface{}
	topicLevelTags          []spec.Tag
	schemaLevelTags         []spec.Tag
	bindings                *bindings
	example                 interface{}
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

func (details *accountDetails) getTags() error {
	// Get topic level tags
	topicLevelTags, _, err := details.srClient.DefaultApi.GetTags(details.srContext, "kafka_topic", details.cluster.Id+":"+details.channelDetails.currentTopic.Name)
	if err != nil {
		return fmt.Errorf("failed to get topic level tags: %v", err)
	}
	for _, topicLevelTag := range topicLevelTags {
		details.channelDetails.topicLevelTags = append(details.channelDetails.topicLevelTags, spec.Tag{Name: topicLevelTag.TypeName})
	}

	// Get schema level tags
	schemaLevelTags, _, err := details.srClient.DefaultApi.GetTags(details.srContext, "sr_schema", strconv.Itoa(int(details.channelDetails.schema.Id)))
	if err != nil {
		return fmt.Errorf("failed to get schema level tags: %v", err)
	}
	for _, schemaLevelTag := range schemaLevelTags {
		details.channelDetails.schemaLevelTags = append(details.channelDetails.schemaLevelTags, spec.Tag{Name: schemaLevelTag.TypeName})
	}

	return nil
}

func (details *accountDetails) getSchemaDetails() error {
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

func (details *accountDetails) getTopicDescription() error {
	atlasEntityWithExtInfo, _, err := details.srClient.DefaultApi.GetByUniqueAttributes(details.srContext, "kafka_topic", details.cluster.Id+":"+details.channelDetails.currentTopic.Name, nil)
	if err != nil {
		return err
	}
	details.channelDetails.currentTopicDescription = atlasEntityWithExtInfo.Entity.Attributes["description"]
	return nil
}

func (details *accountDetails) buildMessageEntity() *spec.MessageEntity {
	entityProducer := new(spec.MessageEntity)
	(*spec.MessageEntity).WithContentType(entityProducer, details.channelDetails.contentType)
	if details.channelDetails.contentType == "application/avro" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/vnd.apache.avro;version=1.9.0")
	} else if details.channelDetails.contentType == "application/json" {
		(*spec.MessageEntity).WithSchemaFormat(entityProducer, "application/schema+json;version=draft-07")
	}
	(*spec.MessageEntity).WithTags(entityProducer, details.channelDetails.schemaLevelTags...)
	// Name
	(*spec.MessageEntity).WithName(entityProducer, msgName(details.channelDetails.currentTopic.Name))
	// Example
	if details.channelDetails.example != nil {
		(*spec.MessageEntity).WithExamples(entityProducer, spec.MessageOneOf1OneOf1ExamplesItems{Payload: &details.channelDetails.example})
	}
	(*spec.MessageEntity).WithBindings(entityProducer, details.channelDetails.bindings.messageBinding)
	(*spec.MessageEntity).WithPayload(entityProducer, details.channelDetails.unmarshalledSchema)
	return entityProducer
}
