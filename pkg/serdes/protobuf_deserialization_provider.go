package serdes

import (
	"fmt"
	"os"
	"regexp"

	"google.golang.org/protobuf/encoding/protojson"
	gproto "google.golang.org/protobuf/proto"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"
)

type ProtobufDeserializationProvider struct {
	deser   *protobuf.Deserializer
	message gproto.Message
}

func (p *ProtobufDeserializationProvider) InitDeserializer(srClientUrl, srClusterId, mode string, srAuth SchemaRegistryAuth, existingClient any) error {
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
		if srAuth.ApiKey != "" && srAuth.ApiSecret != "" {
			serdeClientConfig = schemaregistry.NewConfigWithBasicAuthentication(srClientUrl, srAuth.ApiKey, srAuth.ApiSecret)
		} else if srAuth.Token != "" {
			serdeClientConfig = schemaregistry.NewConfigWithBearerAuthentication(srClientUrl, srAuth.Token, srClusterId, "")
		} else {
			return fmt.Errorf("schema registry client authentication should be provider to initialize deserializer")
		}
		serdeClientConfig.SslCaLocation = srAuth.CertificateAuthorityPath
		serdeClientConfig.SslCertificateLocation = srAuth.ClientCertPath
		serdeClientConfig.SslKeyLocation = srAuth.ClientKeyPath
		serdeClient, err = schemaregistry.NewClient(serdeClientConfig)
	}

	if err != nil {
		return fmt.Errorf("failed to create deserializer-specific Schema Registry client: %w", err)
	}

	serdeConfig := protobuf.NewDeserializerConfig()

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

func (p *ProtobufDeserializationProvider) Deserialize(topic string, headers []kafka.Header, payload []byte) (string, error) {
	// Register the protobuf message
	err := p.deser.ProtoRegistry.RegisterMessage(p.message.ProtoReflect().Type())
	re := regexp.MustCompile(`message .* is already registered`)

	// If the error is due to already registered message, we shouldn't early return
	if err != nil && !re.MatchString(err.Error()) {
		return "", fmt.Errorf("failed to deserialize payload: %w", err)
	}

	// Deserialize the payload into the msgObj
	msgObj, err := p.deser.Deserialize(topic, payload)
	if err != nil {
		return "", fmt.Errorf("failed to deserialize payload: %w", err)
	}

	// Use protojson library to marshal the message to JSON in a compact format
	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,  // Use original field names (snake_case) instead of camelCase
		EmitUnpopulated: false, // Emit unset fields, false here to omit
		Indent:          "",    // No indentation or additional formatting
	}

	jsonBytes, err := marshaler.Marshal(msgObj.(gproto.Message))
	if err != nil {
		return "", fmt.Errorf("failed to convert protobuf message into string after deserialization: %w", err)
	}

	return string(jsonBytes), nil
}
