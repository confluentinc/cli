package serdes

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	gproto "google.golang.org/protobuf/proto"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
)

type ProtobufDeserializationProvider struct {
	deser   *protobuf.Deserializer
	message gproto.Message
}

func (p *ProtobufDeserializationProvider) InitDeserializer(srClientUrl, srClusterId, mode, srApiKey, srApiSecret, token string, existingClient any) error {
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

	serdeConfig := protobuf.NewDeserializerConfig()

	var serdeType serde.Type
	switch mode {
	case "key":
		serdeType = serde.KeySerde
	case "value":
		serdeType = serde.ValueSerde
	default:
		return fmt.Errorf("unknown deserialization mode: %s", mode)
	}

	deser, err := protobuf.NewDeserializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to initialize PROTOBUF deserializer: %w", err)
	}

	p.deser = deser
	return nil
}

func (p *ProtobufDeserializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	message, err := parseMessage(schemaPath, referencePathMap)
	if err != nil {
		return err
	}
	p.message = message
	return nil
}

func (p *ProtobufDeserializationProvider) Deserialize(topic string, payload []byte) (string, error) {
	err := p.deser.DeserializeInto(topic, payload, p.message)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize payload: %w", err)
	}

	// Use protojson to marshal the message to JSON in a compact format
	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,  // Use original field names (snake_case) instead of camelCase
		EmitUnpopulated: false, // Emit unset fields, false here to omit
		Indent:          "",    // No indentation or additional formatting
	}

	jsonBytes, err := marshaler.Marshal(p.message)
	if err != nil {
		return "", fmt.Errorf("failed to convert protobuf message into string after deserialization: %w", err)
	}

	return string(jsonBytes), nil
}
