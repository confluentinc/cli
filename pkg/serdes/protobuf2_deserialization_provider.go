package serdes

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
)

type Protobuf2DeserializationProvider struct {
	deser *protobuf.Deserializer
}

func (a *Protobuf2DeserializationProvider) InitDeserializer(srClientUrl, mode string) error {
	serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create deserializer-specific Schema Registry client: %w", err)
	}

	serdeConfig := protobuf.NewDeserializerConfig()

	var serdeType serde.Type
	switch mode {
	case "key":
		serdeType = serde.KeySerde
	case "value":
		serdeType = serde.ValueSerde
	default:
		return fmt.Errorf("unknown deserialization mode: %s", mode)
	}

	deser, err := protobuf.NewDeserializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to create deserializer: %w", err)
	}

	a.deser = deser
	return nil
}

func (a *Protobuf2DeserializationProvider) Deserialize(topic string, payload []byte, msg interface{}) error {
	err := a.deser.DeserializeInto(topic, payload, msg)
	if err != nil {
		return fmt.Errorf("failed to deserialize payload: %w", err)
	}
	return nil
}
