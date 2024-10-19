package serdes

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/jsonschema"
)

type JsonSerializationProvider struct {
	ser *jsonschema.Serializer
}

func (j *JsonSerializationProvider) InitSerializer(srClientUrl, srClusterId, mode, srApiKey, srApiSecret, token string, schemaId int) error {
	var serdeClientConfig *schemaregistry.Config
	if srApiKey != "" && srApiSecret != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBasicAuthentication(srClientUrl, srApiKey, srApiSecret)
	} else if token != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBearerAuthentication(srClientUrl, token, srClusterId, "")
	} else {
		return fmt.Errorf("schema registry client authentication should be provider to initialize serializer")
	}
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	// Configure the serde settings
	// If schemaId > 0 then use the intended schema ID
	// otherwise use the latest schema ID
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
		return fmt.Errorf("failed to initialize JSON serializer: %w", err)
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
