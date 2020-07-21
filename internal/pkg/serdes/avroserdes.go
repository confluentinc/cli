package serdes

import (
	"io/ioutil"

	"github.com/linkedin/goavro/v2"
)

type AvroSerializationProvider struct {
	codec *goavro.Codec
}

func (avroProvider *AvroSerializationProvider) LoadSchema(schemaPath string) error {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	codec, err := goavro.NewCodec(string(schema))
	if err != nil {
		return err
	}
	avroProvider.codec = codec
	return nil
}

func (avroProvider *AvroSerializationProvider) GetSchemaName() string {
	return AVROSCHEMANAME
}

func (avroProvider *AvroSerializationProvider) encode(str string) ([]byte, error) {
	textual := []byte(str)

	// Convert to native Go object.
	native, _, err := avroProvider.codec.NativeFromTextual(textual)
	if err != nil {
		return nil, err
	}

	binary, err := avroProvider.codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, err
	}
	return binary, nil
}

type AvroDeserializationProvider struct {
	codec *goavro.Codec
}

func (avroProvider *AvroDeserializationProvider) LoadSchema(schemaPath string) error {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	codec, err := goavro.NewCodec(string(schema))
	if err != nil {
		return err
	}
	avroProvider.codec = codec
	return nil
}

func (avroProvider *AvroDeserializationProvider) GetSchemaName() string {
	return AVROSCHEMANAME
}

func (avroProvider *AvroDeserializationProvider) decode(data []byte) (string, error) {
	// Convert to native Go object.
	native, _, err := avroProvider.codec.NativeFromBinary(data)
	if err != nil {
		return "", err
	}

	textual, err := avroProvider.codec.TextualFromNative(nil, native)
	if err != nil {
		return "", err
	}

	return string(textual), nil
}
