package serdes

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
)

type Avro2SerializationProvider struct {
	ser avrov2.Serializer
}

func (a *Avro2SerializationProvider) InitSerializer(srClientUrl, mode string) error {
	serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("Failed to create serializer specific schema registry client: %s\n", err)
	}

	serdeConfig := avrov2.NewSerializerConfig()
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

	ser, err := avrov2.NewSerializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("Failed to create serializer: %s\n", err)
	}

	a.ser = *ser
	return nil
}

func (a *Avro2SerializationProvider) GetSchemaName() string {
	return avroSchemaBackendName
}

func (a *Avro2SerializationProvider) Serialize(topic string, msg interface{}) ([]byte, error) {
	payload, err := a.ser.Serialize(topic, msg)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize payload: %s\n", err)
	}
	return payload, nil
}
