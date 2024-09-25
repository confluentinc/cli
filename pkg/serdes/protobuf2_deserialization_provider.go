package serdes

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
)

type Protobuf2DeserializationProvider struct {
	deser protobuf.Deserializer
}

func (a *Protobuf2DeserializationProvider) InitDeserializer(srClientUrl, mode string) error {
	serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("Failed to create deserializer specific schema registry client: %s\n", err)
	}

	serdeConfig := protobuf.NewDeserializerConfig()

	var serdeType serde.Type
	if mode == "key" {
		serdeType = serde.KeySerde
	} else if mode == "value" {
		serdeType = serde.ValueSerde
	} else {
		return fmt.Errorf("unknown deserialization mode: %s", mode)
	}

	deser, err := protobuf.NewDeserializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("Failed to create deserializer: %s\n", err)
	}

	a.deser = *deser
	return nil
}

func (a *Protobuf2DeserializationProvider) Deserialize(topic string, payload []byte, msg interface{}) error {
	return a.deser.DeserializeInto(topic, payload, msg)
}
