package serdes

import (
	"fmt"

	"github.com/confluentinc/cli/v3/pkg/errors"
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
	avroSchemaName     = "avro"
	doubleSchemaName   = "double"
	integerSchemaName  = "integer"
	jsonSchemaName     = "jsonschema"
	protobufSchemaName = "protobuf"
	stringSchemaName   = "string"
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

type SerializationProvider interface {
	LoadSchema(string, map[string]string) error
	Serialize(string) ([]byte, error)
	GetSchemaName() string
}

type DeserializationProvider interface {
	LoadSchema(string, map[string]string) error
	Deserialize([]byte) (string, error)
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
		return new(JsonSchemaDeserializationProvider), nil
	case protobufSchemaName:
		return new(ProtobufDeserializationProvider), nil
	case stringSchemaName:
		return new(StringDeserializationProvider), nil
	default:
		return nil, fmt.Errorf(errors.UnknownValueFormatErrorMsg)
	}
}
