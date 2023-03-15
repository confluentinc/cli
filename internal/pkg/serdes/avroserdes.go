package serdes

import (
	"os"

	"github.com/linkedin/goavro/v2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type AvroSerializationProvider struct {
	codec *goavro.Codec
}

func (avroProvider *AvroSerializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	if len(referencePathMap) > 0 {
		return errors.New(errors.AvroReferenceNotSupportedErrorMsg)
	}

	schema, err := os.ReadFile(schemaPath)
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
	return AVROSCHEMABACKEND
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

func (avroProvider *AvroDeserializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	if len(referencePathMap) > 0 {
		return errors.New(errors.AvroReferenceNotSupportedErrorMsg)
	}

	schema, err := os.ReadFile(schemaPath)
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
	return AVROSCHEMABACKEND
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
