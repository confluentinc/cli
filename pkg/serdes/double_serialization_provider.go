package serdes

import (
	"encoding/binary"
	"math"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
)

type DoubleSerializationProvider struct{}

func (DoubleSerializationProvider) InitSerializer(_, _, _ string, _ int, _ SchemaRegistryAuth) error {
	return nil
}

func (DoubleSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (DoubleSerializationProvider) Serialize(_, message string) ([]kafka.Header, []byte, error) {
	f, err := strconv.ParseFloat(message, 64)
	if err != nil {
		return nil, nil, err
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(f))

	return nil, buf, nil
}

func (DoubleSerializationProvider) GetSchemaName() string {
	return ""
}

func (DoubleSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return nil
}

func (DoubleSerializationProvider) SetSchemaIDSerializer(_ serde.SchemaIDSerializerFunc) {}
