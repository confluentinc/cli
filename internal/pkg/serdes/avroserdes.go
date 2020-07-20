package serdes

import (
	"io/ioutil"

	"github.com/linkedin/goavro/v2"
)

type AvroSerializationProvider uint32

func (avroProvider *AvroSerializationProvider) GetSchemaName() string {
	return "AVRO"
}

func (avroProvider *AvroSerializationProvider) encode(str string, schemaPath string) ([]byte, error) {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	codec, err := goavro.NewCodec(string(schema))
	if err != nil {
		return nil, err
	}
	textual := []byte(str)

	// Convert to native Go object.
	native, _, err := codec.NativeFromTextual(textual)
	if err != nil {
		return nil, err
	}

	binary, err := codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, err
	}
	return binary, nil
}

type AvroDeserializationProvider uint32

func (avroProvider *AvroDeserializationProvider) GetSchemaName() string {
	return "AVRO"
}

func (avroProvider *AvroDeserializationProvider) decode(data []byte, schemaPath string) (string, error) {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return "", err
	}

	codec, err := goavro.NewCodec(string(schema))
	if err != nil {
		return "", err
	}

	// Convert to native Go object.
	native, _, err := codec.NativeFromBinary(data)
	if err != nil {
		return "", err
	}

	textual, err := codec.TextualFromNative(nil, native)
	if err != nil {
		return "", err
	}

	return string(textual), nil
}
