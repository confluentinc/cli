package serdes

import "github.com/confluentinc/confluent-kafka-go/v2/kafka"

type StringDeserializationProvider struct{}

func (s *StringDeserializationProvider) InitDeserializer(_, _, _ string, _ SchemaRegistryAuth, _ any) error {
	return nil
}

func (s *StringDeserializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *StringDeserializationProvider) Deserialize(_ string, _ []kafka.Header, data []byte) (string, error) {
	message := string(data)
	return message, nil
}
