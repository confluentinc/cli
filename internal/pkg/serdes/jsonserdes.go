package serdes

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/xeipuuv/gojsonschema"
)

type JsonSerializationProvider uint32

func (jsonProvider *JsonSerializationProvider) GetSchemaName() string {
	return "JSON"
}

func (jsonProvider *JsonSerializationProvider) encode(str string, schemaPath string) ([]byte, error) {
	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return nil, err
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schema))
	documentLoader := gojsonschema.NewStringLoader(str)

	// Json schema conducts validation on Json string before serialization.
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, err
	}

	if !result.Valid() {
		return nil, errors.New("The document is not valid.")
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

type JsonDeserializationProvider uint32

func (jsonProvider *JsonDeserializationProvider) GetSchemaName() string {
	return "JSON"
}

func (jsonProvider *JsonDeserializationProvider) decode(data []byte, schemaPath string) (string, error) {
	str := string(data)

	schema, err := ioutil.ReadFile(schemaPath)
	if err != nil {
		return "", err
	}

	schemaLoader := gojsonschema.NewStringLoader(string(schema))
	documentLoader := gojsonschema.NewStringLoader(str)

	// Json schema conducts validation on Json string before serialization.
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return "", err
	}

	if !result.Valid() {
		return "", errors.New("The document is not valid.")
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
