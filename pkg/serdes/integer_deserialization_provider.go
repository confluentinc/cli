package serdes

import (
	"encoding/binary"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
)

type IntegerDeserializationProvider struct{}

func (IntegerDeserializationProvider) InitDeserializer(_, _, _ string, _ SchemaRegistryAuth, _ schemaregistry.Client) error {
	return nil
}

func (IntegerDeserializationProvider) LoadSchema(_ string, _ string, _ serde.Type, _ *kafka.Message) error {
	return nil
}

func (IntegerDeserializationProvider) Deserialize(_ string, _ []kafka.Header, data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	if len(data) != 4 {
		return "", fmt.Errorf("the integer key is invalid")
	}

	message := fmt.Sprintf("%d", binary.LittleEndian.Uint32(data))
	return message, nil
}

func (IntegerDeserializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return nil
}
