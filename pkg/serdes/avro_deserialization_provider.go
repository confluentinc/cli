package serdes

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
)

type AvroDeserializationProvider struct {
	deser *avrov2.Deserializer
}

func (a *AvroDeserializationProvider) InitDeserializer(srClientUrl, srClusterId, mode, srApiKey, srApiSecret, token string, existingClient any) error {
	// Note: Now Serializer/Deserializer are tightly coupled with Schema Registry
	// If existingClient is not nil, we should share this client between ser and deser.
	// As the shared client is referred as mock client to store the same set of schemas in cache
	// If existingClient is nil (which is normal case), ser and deser don't have to share the same client.
	var serdeClient schemaregistry.Client
	var err error
	var ok bool

	if existingClient != nil {
		serdeClient, ok = existingClient.(schemaregistry.Client)
		if !ok {
			return fmt.Errorf("failed to cast existing schema registry client to expected type")
		}
	} else {
		var serdeClientConfig *schemaregistry.Config
		if srApiKey != "" && srApiSecret != "" {
			serdeClientConfig = schemaregistry.NewConfigWithBasicAuthentication(srClientUrl, srApiKey, srApiSecret)
		} else if token != "" {
			serdeClientConfig = schemaregistry.NewConfigWithBearerAuthentication(srClientUrl, token, srClusterId, "")
		} else {
			return fmt.Errorf("schema registry client authentication should be provider to initialize deserializer")
		}
		serdeClient, err = schemaregistry.NewClient(serdeClientConfig)
	}

	if err != nil {
		return fmt.Errorf("failed to create deserializer-specific Schema Registry client: %w", err)
	}

	serdeConfig := avrov2.NewDeserializerConfig()

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

	deser, err := avrov2.NewDeserializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to initialize AVRO deserializer: %w", err)
	}

	a.deser = deser
	return nil
}

func (a *AvroDeserializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (a *AvroDeserializationProvider) Deserialize(topic string, payload []byte) (string, error) {
	message := make(map[string]any)
	err := a.deser.DeserializeInto(topic, payload, &message)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize payload: %w", err)
	}
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to convert generic map message into string after deserialization: %w", err)
	}

	return string(jsonBytes), nil
}
