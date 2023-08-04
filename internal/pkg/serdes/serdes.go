package serdes

import "github.com/confluentinc/cli/internal/pkg/errors"

const (
	AvroSchemaName     string = "avro"
	IntegerSchemaName  string = "integer"
	JsonSchemaName     string = "jsonschema"
	ProtobufSchemaName string = "protobuf"
	StringSchemaName   string = "string"
)

const (
	AvroSchemaBackendName     string = "AVRO"
	JsonSchemaBackendName     string = "JSON"
	ProtobufSchemaBackendName string = "PROTOBUF"
)

type SerializationProvider interface {
	LoadSchema(string, map[string]string) error
	encode(string) ([]byte, error)
	GetSchemaName() string
}

type DeserializationProvider interface {
	LoadSchema(string, map[string]string) error
	decode([]byte) (string, error)
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
		return "", errors.New(errors.UnknownValueFormatErrorMsg)
	}
	return cliValueFormat, nil
}

func GetSerializationProvider(valueFormat string) (SerializationProvider, error) {
	switch valueFormat {
	case AvroSchemaName:
		return new(AvroSerializationProvider), nil
	case IntegerSchemaName:
		return new(IntegerSerializationProvider), nil
	case JsonSchemaName:
		return new(JsonSerializationProvider), nil
	case ProtobufSchemaName:
		return new(ProtobufSerializationProvider), nil
	case StringSchemaName:
		return new(StringSerializationProvider), nil
	default:
		return nil, errors.New(errors.UnknownValueFormatErrorMsg)
	}
}

func GetDeserializationProvider(valueFormat string) (DeserializationProvider, error) {
	switch valueFormat {
	case AvroSchemaName:
		return new(AvroDeserializationProvider), nil
	case IntegerSchemaName:
		return new(IntegerDeserializationProvider), nil
	case JsonSchemaName:
		return new(JsonSchemaDeserializationProvider), nil
	case ProtobufSchemaName:
		return new(ProtobufDeserializationProvider), nil
	case StringSchemaName:
		return new(StringDeserializationProvider), nil
	default:
		return nil, errors.New(errors.UnknownValueFormatErrorMsg)
	}
}

func Serialize(provider SerializationProvider, str string) ([]byte, error) {
	return provider.encode(str)
}

func Deserialize(provider DeserializationProvider, data []byte) (string, error) {
	return provider.decode(data)
}
