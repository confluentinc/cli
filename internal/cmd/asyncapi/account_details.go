package asyncapi

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	ckgo "github.com/confluentinc/confluent-kafka-go/kafka"
	schemaregistry "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/swaggest/go-asyncapi/spec-2.1.0"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type channelDetails struct {
	currentTopic       *schedv1.TopicDescription
	currentSubject     string
	contentType        string
	schema             *schemaregistry.Schema
	unmarshalledSchema map[string]interface{}
	mapOfMessageCompat map[string]interface{}
	tags               []spec.Tag
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
