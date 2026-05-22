package serdes

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
)

type StringSerializationProvider struct{}

func (s *StringSerializationProvider) InitSerializer(_, _, _ string, _ int, _ SchemaRegistryAuth) error {
	return nil
}

func (s *StringSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *StringSerializationProvider) Serialize(_, message string) ([]kafka.Header, []byte, error) {
	return nil, []byte(message), nil
}

func (s *StringSerializationProvider) GetSchemaName() string {
	return ""
}

func (s *StringSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return nil
}

func (s *StringSerializationProvider) SetSchemaIDSerializer(_ serde.SchemaIDSerializerFunc) {}
