package serdes

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
)

type StringDeserializationProvider struct{}

func (s *StringDeserializationProvider) InitDeserializer(_, _, _ string, _ SchemaRegistryAuth, _ schemaregistry.Client) error {
	return nil
}

func (s *StringDeserializationProvider) LoadSchema(_ string, _ string, _ serde.Type, _ *kafka.Message) error {
	return nil
}

func (s *StringDeserializationProvider) Deserialize(_ string, _ []kafka.Header, data []byte) (string, error) {
	message := string(data)
	return message, nil
}

func (s *StringDeserializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return nil
}
