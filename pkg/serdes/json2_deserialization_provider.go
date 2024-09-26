package serdes

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/jsonschema"
)

type Json2DeserializationProvider struct {
	deser *jsonschema.Deserializer
}

func (a *Json2DeserializationProvider) InitDeserializer(srClientUrl, mode string) error {
	serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create deserializer-specific Schema Registry client: %w", err)
	}

	serdeConfig := jsonschema.NewDeserializerConfig()

	var serdeType serde.Type
	switch mode {
	case "key":
		serdeType = serde.KeySerde
	case "value":
		serdeType = serde.ValueSerde
	default:
		return fmt.Errorf("unknown deserialization mode: %s", mode)
	}

	deser, err := jsonschema.NewDeserializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to create deserializer: %w", err)
	}

	a.deser = deser
	return nil
}

func (a *Json2DeserializationProvider) Deserialize(topic string, payload []byte, msg interface{}) error {
	err := a.deser.DeserializeInto(topic, payload, msg)
	if err != nil {
		return fmt.Errorf("failed to deserialize payload: %w", err)
	}
	return nil
}
