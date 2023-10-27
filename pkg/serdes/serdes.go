package serdes

import (
	"fmt"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

const (
	AvroSchemaName     = "avro"
	DoubleSchemaName   = "double"
	IntegerSchemaName  = "integer"
	JsonSchemaName     = "jsonschema"
	ProtobufSchemaName = "protobuf"
	StringSchemaName   = "string"
)

const (
	AvroSchemaBackendName     = "AVRO"
	JsonSchemaBackendName     = "JSON"
	ProtobufSchemaBackendName = "PROTOBUF"
)

var SchemaBasedFormats = []string{"avro", "jsonschema", "protobuf"}

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
	case "", AvroSchemaBackendName:
		cliValueFormat = AvroSchemaName
	case ProtobufSchemaBackendName:
		cliValueFormat = ProtobufSchemaName
	case JsonSchemaBackendName:
		cliValueFormat = JsonSchemaName
	default:
		return "", fmt.Errorf(errors.UnknownValueFormatErrorMsg)
	}
	return cliValueFormat, nil
}

func GetSerializationProvider(valueFormat string) (SerializationProvider, error) {
	switch valueFormat {
	case AvroSchemaName:
		return new(AvroSerializationProvider), nil
	case DoubleSchemaName:
		return new(DoubleSerializationProvider), nil
	case IntegerSchemaName:
		return new(IntegerSerializationProvider), nil
	case JsonSchemaName:
		return new(JsonSerializationProvider), nil
	case ProtobufSchemaName:
		return new(ProtobufSerializationProvider), nil
	case StringSchemaName:
		return new(StringSerializationProvider), nil
	default:
		return nil, fmt.Errorf(errors.UnknownValueFormatErrorMsg)
	}
}

func GetDeserializationProvider(valueFormat string) (DeserializationProvider, error) {
	switch valueFormat {
	case AvroSchemaName:
		return new(AvroDeserializationProvider), nil
	case DoubleSchemaName:
		return new(DoubleDeserializationProvider), nil
	case IntegerSchemaName:
		return new(IntegerDeserializationProvider), nil
	case JsonSchemaName:
		return new(JsonSchemaDeserializationProvider), nil
	case ProtobufSchemaName:
		return new(ProtobufDeserializationProvider), nil
	case StringSchemaName:
		return new(StringDeserializationProvider), nil
	default:
		return nil, fmt.Errorf(errors.UnknownValueFormatErrorMsg)
	}
}
