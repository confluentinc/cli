package serdes

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/confluentinc/cli/internal/pkg/errors"

	"github.com/xeipuuv/gojsonschema"
)

type JsonSerializationProvider struct {
	schemaLoader gojsonschema.JSONLoader
}

func (jsonProvider *JsonSerializationProvider) LoadSchema(schemaPath string) error {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return errors.New(errors.JsonSchemaInvalidErrorMsg)
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schema))
	jsonProvider.schemaLoader = schemaLoader
	return nil
}

func (jsonProvider *JsonSerializationProvider) GetSchemaName() string {
	return JSONSCHEMABACKEND
}

func (jsonProvider *JsonSerializationProvider) encode(str string) ([]byte, error) {
	documentLoader := gojsonschema.NewStringLoader(str)

	// Json schema conducts validation on Json string before serialization.
	result, err := gojsonschema.Validate(jsonProvider.schemaLoader, documentLoader)
	if err != nil {
		return nil, err
	}

	if !result.Valid() {
		return nil, errors.New(errors.JsonDocumentInvalidErrorMsg)
	}

	data := []byte(str)

	// Compact Json string, i.e. remove redundant space, etc.
	compactedBuffer := new(bytes.Buffer)
	err = json.Compact(compactedBuffer, data)
	if err != nil {
		return nil, err
	}
	return compactedBuffer.Bytes(), nil
}

type JsonDeserializationProvider struct {
	schemaLoader gojsonschema.JSONLoader
}

func (jsonProvider *JsonDeserializationProvider) LoadSchema(schemaPath string) error {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return errors.New(errors.JsonSchemaInvalidErrorMsg)
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schema))
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
	result, err := gojsonschema.Validate(jsonProvider.schemaLoader, documentLoader)
	if err != nil {
		return "", err
	}

	if !result.Valid() {
		return "", errors.New(errors.JsonDocumentInvalidErrorMsg)
	}

	data = []byte(str)

	// Compact Json string, i.e. remove redundant space, etc.
	compactedBuffer := new(bytes.Buffer)
	err = json.Compact(compactedBuffer, data)
	if err != nil {
		return "", err
	}
	return string(compactedBuffer.Bytes()), nil
}
