package serdes

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
)

type Protobuf2SerializationProvider struct {
	ser protobuf.Serializer
}

func (a *Protobuf2SerializationProvider) InitSerializer(srClientUrl, mode string) error {
	serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	serdeConfig := protobuf.NewSerializerConfig()
	serdeConfig.AutoRegisterSchemas = false
	serdeConfig.UseLatestVersion = true

	var serdeType serde.Type
	if mode == "key" {
		serdeType = serde.KeySerde
	} else if mode == "value" {
		serdeType = serde.ValueSerde
	} else {
		return fmt.Errorf("unknown serialization mode: %s", mode)
	}

	ser, err := protobuf.NewSerializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer: %w", err)
	}

	a.ser = *ser
	return nil
}

func (a *Protobuf2SerializationProvider) GetSchemaName() string {
	return protobufSchemaBackendName
}

func (a *Protobuf2SerializationProvider) Serialize(topic string, msg interface{}) ([]byte, error) {
	payload, err := a.ser.Serialize(topic, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload: %w", err)
	}
	return payload, nil
}
