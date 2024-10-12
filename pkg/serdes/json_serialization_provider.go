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

func (j *JsonSerializationProvider) InitSerializer(srClientUrl, mode string, schemaId int) error {
	serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	// Configure the serde settings
	// If schemaId > 0 then use the intended schema ID
	// otherwise use the latest schema ID
	// Configuring this correctly determines the underlying serialize strategy
	serdeConfig := jsonschema.NewSerializerConfig()
	serdeConfig.AutoRegisterSchemas = false
	serdeConfig.UseLatestVersion = true
	serdeConfig.EnableValidation = true
	if schemaId > 0 {
		serdeConfig.UseSchemaID = schemaId
		serdeConfig.UseLatestVersion = false
	}

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

	j.ser = ser
	return nil
}

func (j *JsonSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (j *JsonSerializationProvider) GetSchemaName() string {
	return jsonSchemaBackendName
}

func (j *JsonSerializationProvider) Serialize(topic, message string) ([]byte, error) {
	// Convert the plain string message from customer type-in into generic map
	var result map[string]any
	err := json.Unmarshal([]byte(message), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message string into generic map for serialization: %w", err)
	}

	payload, err := j.ser.Serialize(topic, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return payload, nil
}

// GetSchemaRegistryClient This getter function is used in mock testing
// as serializer and deserializer have to share the same SR client instance
func (j *JsonSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return j.ser.Client
}
