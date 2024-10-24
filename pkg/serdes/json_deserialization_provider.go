package serdes

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

type JsonSchemaDeserializationProvider struct {
	schemaLoader *gojsonschema.Schema
}

func (j *JsonSchemaDeserializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	schemaLoader, err := parseSchema(schemaPath, referencePathMap)
	if err != nil {
		return err
	}
	j.schemaLoader = schemaLoader
	return nil
}

func (j *JsonSchemaDeserializationProvider) Deserialize(data []byte) (string, error) {
	str := string(data)

	documentLoader := gojsonschema.NewStringLoader(str)

	// JSON schema conducts validation on JSON string before serialization.
	result, err := j.schemaLoader.Validate(documentLoader)
	if err != nil {
		return "", err
	}

	if !result.Valid() {
		return "", fmt.Errorf(errors.JsonDocumentInvalidErrorMsg)
	}

	data = []byte(str)

	// Compact JSON string, i.e. remove redundant space, etc.
	compactedBuffer := new(bytes.Buffer)
	if err := json.Compact(compactedBuffer, data); err != nil {
		return "", err
	}
	return compactedBuffer.String(), nil
}
