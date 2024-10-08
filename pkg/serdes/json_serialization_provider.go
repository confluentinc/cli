package serdes

import (
	"fmt"

	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/jsonschema"
)

type JsonSerializationProvider struct {
	ser *jsonschema.Serializer
}

func (a *JsonSerializationProvider) InitSerializer(srClientUrl, mode string) error {
	serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	serdeConfig := jsonschema.NewSerializerConfig()
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

	ser, err := jsonschema.NewSerializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer: %w", err)
	}

	a.ser = ser
	return nil
}

func (a *JsonSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (a *JsonSerializationProvider) GetSchemaName() string {
	return jsonSchemaBackendName
}

func (a *JsonSerializationProvider) Serialize(topic string, message any) ([]byte, error) {
	// Convert the plain string message from customer into map
	var result map[string]any
	err := json.Unmarshal([]byte(message.(string)), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message string into map struct for serialization: %w", err)
	}

	payload, err := a.ser.Serialize(topic, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return payload, nil
}
