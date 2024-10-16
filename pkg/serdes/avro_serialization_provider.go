package serdes

import (
	"fmt"

	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
)

type AvroSerializationProvider struct {
	ser *avrov2.Serializer
}

func (a *AvroSerializationProvider) InitSerializer(srClientUrl, mode, srApiKey, srApiSecret string, schemaId int) error {
	serdeClientConfig := schemaregistry.NewConfigWithBasicAuthentication(srClientUrl, srApiKey, srApiSecret)
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	// Configure the serde settings
	// If schemaId > 0 then use the intended schema ID
	// otherwise use the latest schema ID
	serdeConfig := avrov2.NewSerializerConfig()
	serdeConfig.AutoRegisterSchemas = false
	serdeConfig.UseLatestVersion = true
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

	ser, err := avrov2.NewSerializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to initialize AVRO serializer: %w", err)
	}

	a.ser = ser
	return nil
}

func (a *AvroSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (a *AvroSerializationProvider) GetSchemaName() string {
	return avroSchemaBackendName
}

func (a *AvroSerializationProvider) Serialize(topic, message string) ([]byte, error) {
	// Convert the plain string message from customer type-in into generic map
	var result map[string]any
	err := json.Unmarshal([]byte(message), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message string into generic map for serialization: %w", err)
	}

	payload, err := a.ser.Serialize(topic, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return payload, nil
}

// GetSchemaRegistryClient This getter function is used in mock testing
// as serializer and deserializer have to share the same SR client instance
func (a *AvroSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return a.ser.Client
}
