package serdes

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/xeipuuv/gojsonschema"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

type JsonSerializationProvider struct {
	schemaLoader *gojsonschema.Schema
}

func (jsonProvider *JsonSerializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	schemaLoader, err := parseSchema(schemaPath, referencePathMap)
	if err != nil {
		return err
	}
	jsonProvider.schemaLoader = schemaLoader
	return nil
}

func (jsonProvider *JsonSerializationProvider) GetSchemaName() string {
	return JSONSCHEMABACKEND
}

func (jsonProvider *JsonSerializationProvider) encode(str string) ([]byte, error) {
	documentLoader := gojsonschema.NewStringLoader(str)

	// Json schema conducts validation on Json string before serialization.
	result, err := jsonProvider.schemaLoader.Validate(documentLoader)
	if err != nil {
		return nil, err
	}

	if !result.Valid() {
		return nil, errors.New(errors.JsonDocumentInvalidErrorMsg)
	}

	data := []byte(str)

	// Compact Json string, i.e. remove redundant space, etc.
	compactedBuffer := new(bytes.Buffer)
	if err := json.Compact(compactedBuffer, data); err != nil {
		return nil, err
	}
	return compactedBuffer.Bytes(), nil
}

type JsonDeserializationProvider struct {
	schemaLoader *gojsonschema.Schema
}

func (jsonProvider *JsonDeserializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	schemaLoader, err := parseSchema(schemaPath, referencePathMap)
	if err != nil {
		return err
	}
	jsonProvider.schemaLoader = schemaLoader
	return nil
}

func (jsonProvider *JsonDeserializationProvider) GetSchemaName() string {
	return JSONSCHEMABACKEND
}

func (jsonProvider *JsonDeserializationProvider) decode(data []byte) (string, error) {
	str := string(data)

	documentLoader := gojsonschema.NewStringLoader(str)

	// Json schema conducts validation on Json string before serialization.
	result, err := jsonProvider.schemaLoader.Validate(documentLoader)
	if err != nil {
		return "", err
	}

	if !result.Valid() {
		return "", errors.New(errors.JsonDocumentInvalidErrorMsg)
	}

	data = []byte(str)

	// Compact Json string, i.e. remove redundant space, etc.
	compactedBuffer := new(bytes.Buffer)
	if err := json.Compact(compactedBuffer, data); err != nil {
		return "", err
	}
	return compactedBuffer.String(), nil
}

func parseSchema(schemaPath string, referencePathMap map[string]string) (*gojsonschema.Schema, error) {
	sl := gojsonschema.NewSchemaLoader()
	for referenceName, referencePath := range referencePathMap {
		refSchema, err := os.ReadFile(referencePath)
		if err != nil {
			return nil, err
		}
		referenceLoader := gojsonschema.NewStringLoader(string(refSchema))
		if err := sl.AddSchema("/"+referenceName, referenceLoader); err != nil {
			return nil, err
		}
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, errors.New(errors.JsonSchemaInvalidErrorMsg)
	}

	return sl.Compile(gojsonschema.NewStringLoader(string(schema)))
}
