package serdes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

type JsonSerializationProvider struct {
	schemaLoader *gojsonschema.Schema
}

func (j *JsonSerializationProvider) LoadSchema(schemaPath string, referencePathMap map[string]string) error {
	schemaLoader, err := parseSchema(schemaPath, referencePathMap)
	if err != nil {
		return err
	}
	j.schemaLoader = schemaLoader
	return nil
}

func (j *JsonSerializationProvider) GetSchemaName() string {
	return JsonSchemaBackendName
}

func (j *JsonSerializationProvider) Serialize(str string) ([]byte, error) {
	documentLoader := gojsonschema.NewStringLoader(str)

	// JSON schema conducts validation on JSON string before serialization.
	result, err := j.schemaLoader.Validate(documentLoader)
	if err != nil {
		return nil, err
	}

	if !result.Valid() {
		return nil, fmt.Errorf(errors.JsonDocumentInvalidErrorMsg)
	}

	data := []byte(str)

	// Compact JSON string, i.e. remove redundant space, etc.
	compactedBuffer := new(bytes.Buffer)
	if err := json.Compact(compactedBuffer, data); err != nil {
		return nil, err
	}
	return compactedBuffer.Bytes(), nil
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
		return nil, fmt.Errorf("the JSON schema is invalid")
	}

	return sl.Compile(gojsonschema.NewStringLoader(string(schema)))
}
