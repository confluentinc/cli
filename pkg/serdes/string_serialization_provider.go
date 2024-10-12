package serdes

import "github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"

type StringSerializationProvider struct{}

func (s *StringSerializationProvider) InitSerializer(_, _ string, _ int) error {
	return nil
}

func (s *StringSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *StringSerializationProvider) Serialize(_, message string) ([]byte, error) {
	return []byte(message), nil
}

func (s *StringSerializationProvider) GetSchemaName() string {
	return ""
}

func (s *StringSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return nil
}
