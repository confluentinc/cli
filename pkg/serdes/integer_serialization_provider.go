package serdes

import (
	"encoding/binary"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
)

type IntegerSerializationProvider struct{}

func (IntegerSerializationProvider) InitSerializer(_, _, _ string, _ int, _ SchemaRegistryAuth) error {
	return nil
}

func (IntegerSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (IntegerSerializationProvider) Serialize(_, message string) ([]kafka.Header, []byte, error) {
	i, err := strconv.ParseUint(message, 10, 32)
	if err != nil {
		return nil, nil, err
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(i))

	return nil, buf, nil
}

func (IntegerSerializationProvider) GetSchemaName() string {
	return ""
}

func (IntegerSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return nil
}

func (IntegerSerializationProvider) SetSchemaIDSerializer(_ serde.SchemaIDSerializerFunc) {
	return
}
