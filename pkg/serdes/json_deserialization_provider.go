package serdes

import (
	"fmt"

	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/jsonschema"
)

type JsonDeserializationProvider struct {
	deser *jsonschema.Deserializer
}

func (j *JsonDeserializationProvider) InitDeserializer(srClientUrl, mode string, existingClient any) error {
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
		serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
		serdeClient, err = schemaregistry.NewClient(serdeClientConfig)
	}

	if err != nil {
		return fmt.Errorf("failed to create deserializer-specific Schema Registry client: %w", err)
	}

	serdeConfig := jsonschema.NewDeserializerConfig()
	serdeConfig.EnableValidation = true

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

func (j *JsonDeserializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (j *JsonDeserializationProvider) Deserialize(topic string, payload []byte) (string, error) {
	message := make(map[string]interface{})
	err := j.deser.DeserializeInto(topic, payload, &message)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize payload: %w", err)
	}
	jsonBytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to convert generic map message into string after deserialization: %w", err)
	}

	return string(jsonBytes), nil
}
