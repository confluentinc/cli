package serdes

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/jsonschema"
)

type JsonDeserializationProvider struct {
	deser *jsonschema.Deserializer
}

func (j *JsonDeserializationProvider) InitDeserializer(srClientUrl, srClusterId, mode string, srAuth SchemaRegistryAuth, existingClient schemaregistry.Client) error {
	// Note: Now Serializer/Deserializer are tightly coupled with Schema Registry
	// If existingClient is not nil, we should share this client between ser and deser.
	// As the shared client is referred as mock client to store the same set of schemas in cache
	// If existingClient is nil (which is normal case), ser and deser don't have to share the same client.
	serdeClient, err := initSchemaRegistryClient(srClientUrl, srClusterId, srAuth, existingClient)
	if err != nil {
		return fmt.Errorf("failed to create deserializer-specific Schema Registry client: %w", err)
	}

	// Note: the EnableValidation = true option has been removed as it is bugged in the JSON deserializer,
	// and also because we don't actually need to validate in the deserializer (only in the serializer)
	serdeConfig := jsonschema.NewDeserializerConfig()

	// local KMS secret is only set and used during local testing with ruleSet
	if localKmsSecretValue := os.Getenv(localKmsSecretMacro); srClientUrl == mockClientUrl && localKmsSecretValue != "" {
		serdeConfig.RuleConfig = map[string]string{
			localKmsSecretKey: localKmsSecretValue,
		}
	}

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
		return fmt.Errorf("failed to initialize JSON deserializer: %w", err)
	}

	j.deser = deser
	return nil
}

func (j *JsonDeserializationProvider) LoadSchema(_ string, _ string, _ serde.Type, _ *kafka.Message) error {
	return nil
}

func (j *JsonDeserializationProvider) Deserialize(topic string, headers []kafka.Header, payload []byte) (string, error) {
	message := make(map[string]interface{})
	if len(headers) > 0 {
		err := j.deser.DeserializeWithHeadersInto(topic, headers, payload, &message)
		if err != nil {
			return "", fmt.Errorf("failed to deserialize payload: %w", err)
		}
	} else {
		err := j.deser.DeserializeInto(topic, payload, &message)
		if err != nil {
			return "", fmt.Errorf("failed to deserialize payload: %w", err)
		}
	}
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to convert generic map message into string after deserialization: %w", err)
	}

	return string(jsonBytes), nil
}

func (j *JsonDeserializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return j.deser.Client
}
