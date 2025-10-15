package serdes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	gproto "google.golang.org/protobuf/proto"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/protobuf"

	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/utils"
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
			serdeClientConfig = schemaregistry.NewConfig(srClientUrl)
			log.CliLogger.Info("initializing deserializer with no schema registry client authentication")
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

func (p *ProtobufDeserializationProvider) LoadSchema(subject, schemaPath string, serdeType serde.Type, kafkaMessage *kafka.Message) error {
	updatedSchemaPath, referencePathMap, err := p.requestSchema(subject, schemaPath, serdeType, kafkaMessage)
	if err != nil {
		return err
	}

	message, err := parseMessage(updatedSchemaPath, referencePathMap)
	if err != nil {
		return err
	}
	p.message = message
	return nil
}

func (p *ProtobufDeserializationProvider) requestSchema(subject, schemaPath string, serdeType serde.Type, message *kafka.Message) (string, map[string]string, error) {
	if message == nil {
		return "", nil, fmt.Errorf("kafka message is nil")
	}

	schemaID := serde.SchemaID{}
	_, err := serde.DualSchemaIDDeserializer(subject, serdeType, message.Headers, message.Value, &schemaID)
	if err != nil {
		return "", nil, err
	}

	var idString string
	if schemaID.ID > 0 { // integer schema ID from the message prefix
		idString = strconv.Itoa(schemaID.ID)
	} else if schemaID.GUID != uuid.Nil { // GUID schema ID from the header
		idString = schemaID.GUID.String()
	}

	tempStorePath := filepath.Join(schemaPath, fmt.Sprintf("%s-%s.txt", subject, idString))
	tempRefStorePath := filepath.Join(schemaPath, fmt.Sprintf("%s-%s.ref", subject, idString))
	var references []schemaregistry.Reference
	if !utils.FileExists(tempStorePath) || !utils.FileExists(tempRefStorePath) {
		var schemaInfo schemaregistry.SchemaInfo
		if schemaID.ID > 0 {
			schemaInfo, err = p.deser.Client.GetBySubjectAndID(subject, schemaID.ID)
			if err != nil {
				return "", nil, err
			}
		} else if schemaID.GUID != uuid.Nil {
			schemaInfo, err = p.deser.Client.GetByGUID(schemaID.GUID.String())
			if err != nil {
				return "", nil, err
			}
		}

		if err := os.WriteFile(tempStorePath, []byte(schemaInfo.Schema), 0644); err != nil {
			return "", nil, err
		}

		refBytes, err := json.Marshal(schemaInfo.References)
		if err != nil {
			return "", nil, err
		}
		if err := os.WriteFile(tempRefStorePath, refBytes, 0644); err != nil {
			return "", nil, err
		}
		references = schemaInfo.References
	} else {
		refBytes, err := os.ReadFile(tempRefStorePath)
		if err != nil {
			return "", nil, err
		}
		if err := json.Unmarshal(refBytes, &references); err != nil {
			return "", nil, err
		}
	}

	referencePathMap := map[string]string{}
	for _, ref := range references {
		refTempStorePath := filepath.Join(schemaPath, ref.Name)
		if !utils.FileExists(refTempStorePath) {
			schemaMetadata, err := p.deser.Client.GetSchemaMetadata(ref.Subject, ref.Version)
			if err != nil {
				return "", nil, err
			}

			var refSchemaInfo schemaregistry.SchemaInfo
			if schemaMetadata.ID > 0 {
				refSchemaInfo, err = p.deser.Client.GetBySubjectAndID(ref.Subject, schemaMetadata.ID)
				if err != nil {
					return "", nil, err
				}
			} else if schemaMetadata.GUID != "" {
				refSchemaInfo, err = p.deser.Client.GetByGUID(schemaMetadata.GUID)
				if err != nil {
					return "", nil, err
				}
			}

			if err := os.MkdirAll(filepath.Dir(refTempStorePath), 0755); err != nil {
				return "", nil, err
			}
			if err := os.WriteFile(refTempStorePath, []byte(refSchemaInfo.Schema), 0644); err != nil {
				return "", nil, err
			}
		}
		referencePathMap[ref.Name] = refTempStorePath
	}

	return tempStorePath, referencePathMap, nil
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
	var msgObj interface{}
	if len(headers) > 0 {
		msgObj, err = p.deser.DeserializeWithHeaders(topic, headers, payload)
		if err != nil {
			return "", fmt.Errorf("failed to deserialize payload: %w", err)
		}
	} else {
		msgObj, err = p.deser.Deserialize(topic, payload)
		if err != nil {
			return "", fmt.Errorf("failed to deserialize payload: %w", err)
		}
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
