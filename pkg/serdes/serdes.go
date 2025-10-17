package serdes

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/log"
)

var DekAlgorithms = []string{
	"AES128_GCM",
	"AES256_GCM",
	"AES256_SIV",
}

var KmsTypes = []string{
	"aws-kms",
	"azure-kms",
	"gcp-kms",
}

const (
	avroSchemaName             = "avro"
	doubleSchemaName           = "double"
	integerSchemaName          = "integer"
	jsonSchemaName             = "jsonschema"
	protobufSchemaName         = "protobuf"
	stringSchemaName           = "string"
	mockClientUrl              = "mock://"
	localKmsSecretKey          = "secret"
	localKmsSecretValueDefault = "default_local_kms_secret_12345"
	localKmsSecretMacro        = "LOCAL_KMS_SECRET"
)

var Formats = []string{
	stringSchemaName,
	avroSchemaName,
	doubleSchemaName,
	integerSchemaName,
	jsonSchemaName,
	protobufSchemaName,
}

var SchemaBasedFormats = []string{
	avroSchemaName,
	jsonSchemaName,
	protobufSchemaName,
}

const (
	avroSchemaBackendName     = "AVRO"
	jsonSchemaBackendName     = "JSON"
	protobufSchemaBackendName = "PROTOBUF"
)

// Struct to hold strings/filepaths relating to Schema Registry authentication or authorization
type SchemaRegistryAuth struct {
	ApiKey                   string
	ApiSecret                string
	CertificateAuthorityPath string
	ClientCertPath           string
	ClientKeyPath            string
	Token                    string
}

type SerializationProvider interface {
	InitSerializer(srClientUrl, srClusterId, mode string, schemaId int, srAuth SchemaRegistryAuth) error
	LoadSchema(string, map[string]string) error
	Serialize(string, string) ([]kafka.Header, []byte, error)
	GetSchemaName() string
	GetSchemaRegistryClient() schemaregistry.Client
	SetSchemaIDSerializer(headerSerializer serde.SchemaIDSerializerFunc) // For unit testing purposes
}

type DeserializationProvider interface {
	InitDeserializer(srClientUrl, srClusterId, mode string, srAuth SchemaRegistryAuth, existingClient schemaregistry.Client) error
	LoadSchema(string, string, serde.Type, *kafka.Message) error
	Deserialize(string, []kafka.Header, []byte) (string, error)
	GetSchemaRegistryClient() schemaregistry.Client
}

func FormatTranslation(backendValueFormat string) (string, error) {
	var cliValueFormat string
	switch backendValueFormat {
	case "", avroSchemaBackendName:
		cliValueFormat = avroSchemaName
	case protobufSchemaBackendName:
		cliValueFormat = protobufSchemaName
	case jsonSchemaBackendName:
		cliValueFormat = jsonSchemaName
	default:
		return "", fmt.Errorf(errors.UnknownValueFormatErrorMsg)
	}
	return cliValueFormat, nil
}

func GetSerializationProvider(valueFormat string) (SerializationProvider, error) {
	switch valueFormat {
	case avroSchemaName:
		return new(AvroSerializationProvider), nil
	case doubleSchemaName:
		return new(DoubleSerializationProvider), nil
	case integerSchemaName:
		return new(IntegerSerializationProvider), nil
	case jsonSchemaName:
		return new(JsonSerializationProvider), nil
	case protobufSchemaName:
		return new(ProtobufSerializationProvider), nil
	case stringSchemaName:
		return new(StringSerializationProvider), nil
	default:
		return nil, fmt.Errorf(errors.UnknownValueFormatErrorMsg)
	}
}

func GetDeserializationProvider(valueFormat string) (DeserializationProvider, error) {
	switch valueFormat {
	case avroSchemaName:
		return new(AvroDeserializationProvider), nil
	case doubleSchemaName:
		return new(DoubleDeserializationProvider), nil
	case integerSchemaName:
		return new(IntegerDeserializationProvider), nil
	case jsonSchemaName:
		return new(JsonDeserializationProvider), nil
	case protobufSchemaName:
		return new(ProtobufDeserializationProvider), nil
	case stringSchemaName:
		return new(StringDeserializationProvider), nil
	default:
		return nil, fmt.Errorf(errors.UnknownValueFormatErrorMsg)
	}
}

func IsProtobufSchema(valueFormat string) bool {
	return valueFormat == protobufSchemaName
}

func initSchemaRegistryClient(srClientUrl, srClusterId string, srAuth SchemaRegistryAuth, existingClient schemaregistry.Client) (schemaregistry.Client, error) {
	if existingClient != nil {
		return existingClient, nil
	}

	var serdeClientConfig *schemaregistry.Config
	if srAuth.ApiKey != "" && srAuth.ApiSecret != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBasicAuthentication(srClientUrl, srAuth.ApiKey, srAuth.ApiSecret)
	} else if srAuth.Token != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBearerAuthentication(srClientUrl, srAuth.Token, srClusterId, "")
	} else {
		serdeClientConfig = schemaregistry.NewConfig(srClientUrl)
		log.CliLogger.Info("initializing schema registry client with no authentication")
	}
	serdeClientConfig.SslCaLocation = srAuth.CertificateAuthorityPath
	serdeClientConfig.SslCertificateLocation = srAuth.ClientCertPath
	serdeClientConfig.SslKeyLocation = srAuth.ClientKeyPath

	return schemaregistry.NewClient(serdeClientConfig)
}
