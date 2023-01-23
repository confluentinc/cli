package serdes

import "github.com/confluentinc/cli/internal/pkg/errors"

const (
	AVROSCHEMANAME        string = "avro"
	AVROSCHEMABACKEND     string = "AVRO"
	JSONSCHEMANAME        string = "jsonschema"
	JSONSCHEMABACKEND     string = "JSON"
	PROTOBUFSCHEMANAME    string = "protobuf"
	PROTOBUFSCHEMABACKEND string = "PROTOBUF"
	RAWSCHEMANAME         string = "string"
)

func FormatTranslation(backendValueFormat string) (string, error) {
	var cliValueFormat string
	switch backendValueFormat {
	case "", AVROSCHEMABACKEND:
		cliValueFormat = AVROSCHEMANAME
	case PROTOBUFSCHEMABACKEND:
		cliValueFormat = PROTOBUFSCHEMANAME
	case JSONSCHEMABACKEND:
		cliValueFormat = JSONSCHEMANAME
	default:
		return "", errors.New(errors.UnknownValueFormatErrorMsg)
	}
	return cliValueFormat, nil
}

func GetSerializationProvider(valueFormat string) (SerializationProvider, error) {
	var provider SerializationProvider
	// Will add other providers in later commits.
	switch valueFormat {
	case AVROSCHEMANAME:
		provider = new(AvroSerializationProvider)
	case PROTOBUFSCHEMANAME:
		provider = new(ProtoSerializationProvider)
	case JSONSCHEMANAME:
		provider = new(JsonSerializationProvider)
	case RAWSCHEMANAME:
		provider = new(RawSerializationProvider)
	default:
		return nil, errors.New(errors.UnknownValueFormatErrorMsg)
	}
	return provider, nil
}

func GetDeserializationProvider(valueFormat string) (DeserializationProvider, error) {
	var provider DeserializationProvider
	// Will add other providers in later commits.
	switch valueFormat {
	case AVROSCHEMANAME:
		provider = new(AvroDeserializationProvider)
	case PROTOBUFSCHEMANAME:
		provider = new(ProtoDeserializationProvider)
	case JSONSCHEMANAME:
		provider = new(JsonDeserializationProvider)
	case RAWSCHEMANAME:
		provider = new(RawDeserializationProvider)
	default:
		return nil, errors.New(errors.UnknownValueFormatErrorMsg)
	}
	return provider, nil
}

type SerializationProvider interface {
	LoadSchema(string, map[string]string) error
	encode(string) ([]byte, error)
	GetSchemaName() string
}

func Serialize(provider SerializationProvider, str string) ([]byte, error) {
	return provider.encode(str)
}

type DeserializationProvider interface {
	LoadSchema(string, map[string]string) error
	decode([]byte) (string, error)
	GetSchemaName() string
}

func Deserialize(provider DeserializationProvider, data []byte) (string, error) {
	return provider.decode(data)
}
