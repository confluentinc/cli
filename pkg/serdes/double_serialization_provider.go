package serdes

import (
	"encoding/binary"
	"math"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
)

type DoubleSerializationProvider struct{}

func (DoubleSerializationProvider) InitSerializer(_, _, _ string, _ int, _ SchemaRegistryAuth) error {
	return nil
}

func (DoubleSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (DoubleSerializationProvider) Serialize(_, message string) ([]byte, error) {
	f, err := strconv.ParseFloat(message, 64)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(f))

	return buf, nil
}

func (DoubleSerializationProvider) GetSchemaName() string {
	return ""
}

func (DoubleSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return nil
}
