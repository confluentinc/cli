package serdes

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
)

type DoubleDeserializationProvider struct{}

func (DoubleDeserializationProvider) InitDeserializer(_, _, _ string, _ SchemaRegistryAuth, _ any) error {
	return nil
}

func (DoubleDeserializationProvider) LoadSchema(_ string, _ string, _ serde.Type, _ *kafka.Message) error {
	return nil
}

func (DoubleDeserializationProvider) Deserialize(_ string, _ []kafka.Header, data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	if len(data) != 8 {
		return "", fmt.Errorf("the double key is invalid")
	}

	message := fmt.Sprintf("%f", math.Float64frombits(binary.LittleEndian.Uint64(data)))
	return message, nil
}
