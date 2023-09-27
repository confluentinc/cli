package serdes

import (
	"os"

	"github.com/linkedin/goavro/v2"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

type AvroSerializationProvider struct {
	codec *goavro.Codec
}

func (a *AvroSerializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
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

func (a *AvroSerializationProvider) GetSchemaName() string {
	return AvroSchemaBackendName
}

func (a *AvroSerializationProvider) SchemaBased() bool {
	return true
}

func (a *AvroSerializationProvider) Serialize(str string) ([]byte, error) {
	textual := []byte(str)

	// Convert to native Go object.
	native, _, err := a.codec.NativeFromTextual(textual)
	if err != nil {
		return nil, err
	}

	binary, err := a.codec.BinaryFromNative(nil, native)
	if err != nil {
		return nil, err
	}
	return binary, nil
}
