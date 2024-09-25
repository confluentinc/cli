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
	InitSerializer(string, string) error
	Serialize(string, interface{}) ([]byte, error)
	GetSchemaName() string
}

type DeserializationProvider interface {
	InitDeserializer(string, string) error
	Deserialize(string, []byte, interface{}) error
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
		return new(Avro2SerializationProvider), nil
	case jsonSchemaName:
		return new(Json2SerializationProvider), nil
	case protobufSchemaName:
		return new(Protobuf2SerializationProvider), nil
	default:
		return nil, fmt.Errorf(errors.UnknownValueFormatErrorMsg)
	}
}

func GetDeserializationProvider(valueFormat string) (DeserializationProvider, error) {
	switch valueFormat {
	case avroSchemaName:
		return new(Avro2DeserializationProvider), nil
	case jsonSchemaName:
		return new(Json2DeserializationProvider), nil
	case protobufSchemaName:
		return new(Protobuf2DeserializationProvider), nil
	default:
		return nil, fmt.Errorf(errors.UnknownValueFormatErrorMsg)
	}
}
