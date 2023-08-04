package serdes

import (
	"os"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/linkedin/goavro/v2"
)

type AvroDeserializationProvider struct {
	codec *goavro.Codec
}

func (a *AvroDeserializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
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
	a.codec = codec
	return nil
}

func (a *AvroDeserializationProvider) decode(data []byte) (string, error) {
	// Convert to native Go object.
	native, _, err := a.codec.NativeFromBinary(data)
	if err != nil {
		return "", err
	}

	textual, err := a.codec.TextualFromNative(nil, native)
	if err != nil {
		return "", err
	}

	return string(textual), nil
}
