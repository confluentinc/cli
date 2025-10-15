package serdes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
)

var tempDir string

func TestMain(m *testing.M) {
	// Create the temporary directory used for placing schemas
	tempDir, _ = createTempDir()

	// Run the required test(s)
	code := m.Run()
	// Sleep for 2s as the Windows usually holds the file(s) access longer
	time.Sleep(2 * time.Second)

	// Cleanup directory here will always run after all test(s) have been executed.
	// Skip this for now on Windows because a .proto file is sometimes locked here in Windows
	if runtime.GOOS != "windows" {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("failed to remove temporary directory: %s", err)
			code = 1
		}
	}

	os.Exit(code)
}

func TestGetSerializationProvider(t *testing.T) {
	req := require.New(t)
	valueFormats := []string{avroSchemaName, jsonSchemaName, protobufSchemaName}
	schemaNames := []string{avroSchemaBackendName, jsonSchemaBackendName, protobufSchemaBackendName}

	for idx, valueFormat := range valueFormats {
		provider, err := GetSerializationProvider(valueFormat)
		req.Nil(err)
		req.Equal(provider.GetSchemaName(), schemaNames[idx])
		req.Nil(err)
	}

	provider, err := GetSerializationProvider("UNKNOWN")
	req.Nil(provider)
	req.EqualError(err, "unknown value schema format")
}

func TestGetDeserializationProvider(t *testing.T) {
	valueFormats := []string{avroSchemaName, protobufSchemaName, jsonSchemaName, stringSchemaName}

	for _, valueFormat := range valueFormats {
		_, err := GetDeserializationProvider(valueFormat)
		require.NoError(t, err)
	}

	_, err := GetDeserializationProvider("UNKNOWN")
	require.Error(t, err)
}

func TestStringSerdes(t *testing.T) {
	req := require.New(t)

	serializationProvider, _ := GetSerializationProvider(stringSchemaName)
	expectedBytes := []byte{115, 111, 109, 101, 83, 116, 114, 105, 110, 103}
	_, data, err := serializationProvider.Serialize("", "someString")
	req.Nil(err)
	result := bytes.Compare(data, expectedBytes)
	req.Zero(result)

	deserializationProvider, _ := GetDeserializationProvider(stringSchemaName)
	data = []byte{115, 111, 109, 101, 83, 116, 114, 105, 110, 103}
	str, err := deserializationProvider.Deserialize("", nil, data)
	req.Nil(err)
	req.Equal(str, "someString")
}

func TestAvroSerdesValid(t *testing.T) {
	req := require.New(t)

	schemaString := `{"type":"record","name":"myRecord","fields":[{"name":"f1","type":"int"}]}`
	schemaPath := filepath.Join(tempDir, "avro-schema.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// expectedBytes[0] is the magic byte, expectedBytes[1:5] is the schemaId in BigIndian
	expectedString := `{"f1":123}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 246, 1}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "AVRO",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(avroSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", nil, data)
	req.Nil(err)

	req.Equal(expectedString, actualString)
}

func TestAvroSerdesValidWithHeaders(t *testing.T) {
	req := require.New(t)

	schemaString := `{"type":"record","name":"myRecord","fields":[{"name":"f1","type":"int"}]}`
	schemaPath := filepath.Join(tempDir, "avro-schema.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":123}`
	expectedBytes := []byte{246, 1}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	serializationProvider.SetSchemaIDSerializer(serde.HeaderSchemaIDSerializer)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "AVRO",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	headers, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(avroSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", headers, data)
	req.Nil(err)

	req.Equal(expectedString, actualString)
}

func TestAvroSerdesInvalid(t *testing.T) {
	req := require.New(t)

	schemaString := `{"type":"record","name":"myRecord","fields":[{"name":"f1","type":"string"}]}`
	schemaPath := filepath.Join(tempDir, "avro-schema-invalid.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "AVRO",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(avroSchemaName)
	req.Nil(err)

	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	brokenString := `{"f1"`
	brokenBytes := []byte{0, 0, 0, 0, 1, 6, 97}

	_, _, err = serializationProvider.Serialize("topic1", brokenString)
	req.Regexp(`cannot decode textual record "myRecord": short buffer`, err)

	_, err = deserializationProvider.Deserialize("topic1", nil, brokenBytes)
	req.Regexp("unexpected EOF$", err)

	invalidString := `{"f2": "abc"}`
	_, _, err = serializationProvider.Serialize("topic1", invalidString)
	req.Regexp(`cannot decode textual map: cannot determine codec: "f2"$`, err)
}

func TestAvroSerdesNestedValid(t *testing.T) {
	req := require.New(t)

	schemaString := `{"type":"record","name":"myRecord","fields":[{"name":"f1","type":"string"},{"name":"nestedField","type":{"type":"record","name":"AnotherNestedRecord","fields":[{"name":"nested_field1","type":"float"},{"name":"nested_field2","type":"string"}]}}]}`
	schemaPath := filepath.Join(tempDir, "avro-schema-nested.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// expectedBytes[0] is the magic byte, expectedBytes[1:5] is the schemaId in BigIndian
	expectedString := `{"f1":"asd","nestedField":{"nested_field1":123.01,"nested_field2":"example"}}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 6, 97, 115, 100, 31, 5, 246, 66, 14, 101, 120, 97, 109, 112, 108, 101}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "AVRO",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(avroSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", nil, expectedBytes)
	req.Nil(err)

	req.Equal(expectedString, actualString)
}

func TestAvroSerdesValidWithRuleSet(t *testing.T) {
	req := require.New(t)

	t.Setenv(localKmsSecretMacro, localKmsSecretValueDefault)

	schemaString := `{"type":"record","name":"myRecord","fields":[{"name":"f1","type":"string","confluent:tags": ["PII"]}]}`
	schemaPath := filepath.Join(tempDir, "avro-schema-ruleset.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":"this is a confidential message in AVRO schema"}`

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// CSFLE specific rules during schema registration
	encRule := schemaregistry.Rule{
		Name: "avro-encrypt",
		Kind: "TRANSFORM",
		Mode: "WRITEREAD",
		Type: "ENCRYPT",
		Tags: []string{"PII"},
		Params: map[string]string{
			"encrypt.kek.name":   "kek1",
			"encrypt.kms.type":   "local-kms",
			"encrypt.kms.key.id": "mykey",
		},
		OnFailure: "ERROR,NONE",
	}
	ruleSet := schemaregistry.RuleSet{
		DomainRules: []schemaregistry.Rule{encRule},
	}

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "AVRO",
		RuleSet:    &ruleSet,
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(avroSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", nil, data)
	req.Nil(err)

	req.Equal(expectedString, actualString)
}

func TestJsonSerdesValid(t *testing.T) {
	req := require.New(t)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string"}},"required":["f1"]}`
	schemaPath := filepath.Join(tempDir, "json-schema.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 125}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "JSON",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	err = deserializationProvider.LoadSchema("topic1-value", schemaPath, serde.ValueSerde, nil)
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", nil, expectedBytes)
	req.Nil(err)
	req.Equal(expectedString, actualString)
}

func TestJsonSerdesValidWithHeaders(t *testing.T) {
	req := require.New(t)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string"}},"required":["f1"]}`
	schemaPath := filepath.Join(tempDir, "json-schema.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 125}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	serializationProvider.SetSchemaIDSerializer(serde.HeaderSchemaIDSerializer)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "JSON",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	headers, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", headers, expectedBytes)
	req.Nil(err)
	req.Equal(expectedString, actualString)
}

func TestJsonSerdesReference(t *testing.T) {
	req := require.New(t)

	// Reference schema should be registered from user side prior to be used as reference
	// So subject and schema version will be known value at this time
	referenceContent := `[{"name":"RefSchema","subject":"topic2-value","version":1}]`
	referenceString := `{"type": "string"}`
	referencePath := filepath.Join(tempDir, "json-reference.json")
	req.NoError(os.WriteFile(referencePath, []byte(referenceContent), 0644))

	// Prepare main schema information
	schemaString := `{"type":"object","properties":{"f1":{"$ref":"RefSchema"}},"required":["f1"]}`
	schemaPath := filepath.Join(tempDir, "json-schema-main.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// Read references from local reference schema file
	references, err := readSchemaReferences(referencePath)
	req.Nil(err)

	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{0, 0, 0, 0, 2, 123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 125}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the reference schema and root schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	referenceInfo := schemaregistry.SchemaInfo{
		Schema:     referenceString,
		SchemaType: "JSON",
	}
	_, err = client.Register("topic2-value", referenceInfo, false)
	req.Nil(err)

	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "JSON",
		References: references,
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	err = deserializationProvider.LoadSchema("topic1-value", schemaPath, serde.ValueSerde, nil)
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", nil, expectedBytes)
	req.Nil(err)
	req.Equal(expectedString, actualString)
}

func TestJsonSerdesInvalid(t *testing.T) {
	req := require.New(t)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string"}},"required":["f1"]}`
	schemaPath := filepath.Join(tempDir, "json-schema-invalid.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "JSON",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	brokenString := `{"f1":`
	brokenBytes := []byte{123, 34, 102, 50}

	_, _, err = serializationProvider.Serialize("topic1", brokenString)
	req.Regexp("unexpected end of JSON input$", err)

	_, err = deserializationProvider.Deserialize("topic1", nil, brokenBytes)
	req.Regexp("unknown magic byte[\\s\\d]*$", err)

	invalidString := `{"f2": "abc"}`
	invalidBytes := []byte{123, 34, 102, 50, 34, 58, 34, 97, 115, 100, 34, 125}

	_, _, err = serializationProvider.Serialize("topic1", invalidString)
	req.Regexp("missing properties: 'f1'$", err)

	_, err = deserializationProvider.Deserialize("topic1", nil, invalidBytes)
	req.Regexp("unknown magic byte[\\s\\d]*$", err)
}

func TestJsonSerdesNestedValid(t *testing.T) {
	req := require.New(t)

	schemaString := `{"type":"object","name":"myRecord","properties":{"f1":{"type":"string"},"nestedField":{"type":"object","name":"AnotherNestedRecord","properties":{"nested_field1":{"type":"number"},"nested_field2":{"type":"string"}}}},"required":["f1","nestedField"]}`
	schemaPath := filepath.Join(tempDir, "json-schema-nested.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// expectedBytes[0] is the magic byte, expectedBytes[1:5] is the schemaId in BigIndian
	expectedString := `{"f1":"asd","nestedField":{"nested_field1":123.01,"nested_field2":"example"}}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 44, 34, 110, 101, 115, 116, 101, 100, 70, 105, 101, 108, 100, 34,
		58, 123, 34, 110, 101, 115, 116, 101, 100, 95, 102, 105, 101, 108, 100, 49, 34, 58, 49, 50, 51, 46, 48, 49, 44, 34, 110, 101, 115, 116, 101, 100, 95,
		102, 105, 101, 108, 100, 50, 34, 58, 34, 101, 120, 97, 109, 112, 108, 101, 34, 125, 125}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "JSON",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", nil, expectedBytes)
	req.Nil(err)

	req.Equal(expectedString, actualString)
}

func TestJsonSerdesValidWithRuleSet(t *testing.T) {
	req := require.New(t)
	t.Setenv(localKmsSecretMacro, localKmsSecretValueDefault)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string","confluent:tags": ["PII"]}},"required":["f1"]}`
	schemaPath := filepath.Join(tempDir, "json-schema-ruleset.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":"this is a confidential message in JSON schema"}`

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err := serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// CSFLE specific rules during schema registration
	encRule := schemaregistry.Rule{
		Name: "json-encrypt",
		Kind: "TRANSFORM",
		Mode: "WRITEREAD",
		Type: "ENCRYPT",
		Tags: []string{"PII"},
		Params: map[string]string{
			"encrypt.kek.name":   "kek1",
			"encrypt.kms.type":   "local-kms",
			"encrypt.kms.key.id": "mykey",
		},
		OnFailure: "ERROR,NONE",
	}
	ruleSet := schemaregistry.RuleSet{
		DomainRules: []schemaregistry.Rule{encRule},
	}

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "JSON",
		RuleSet:    &ruleSet,
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", nil, data)
	req.Nil(err)

	req.Equal(expectedString, actualString)
}

func TestProtobufSerdesValid(t *testing.T) {
	req := require.New(t)

	tempDir, err := os.MkdirTemp(tempDir, "protobuf")
	req.NoError(err)
	defer os.RemoveAll(tempDir)

	schemaString := `
	syntax = "proto3";
	message Person {
	  string name = 1;
	  int32 page = 2;
	  double result = 3;
	}`
	schemaPath := filepath.Join(tempDir, "person-schema.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","page":1,"result":2.5}`

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "PROTOBUF",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema("topic1-value", tempDir, serde.ValueSerde, &kafka.Message{Value: data})
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", nil, data)
	req.Nil(err)
	req.JSONEq(expectedString, actualString)
}

func TestProtobufSerdesValidWithHeaders(t *testing.T) {
	req := require.New(t)

	tempDir, err := os.MkdirTemp(tempDir, "protobuf")
	req.NoError(err)
	defer os.RemoveAll(tempDir)

	schemaString := `
	syntax = "proto3";
	message Person {
	  string name = 1;
	  int32 page = 2;
	  double result = 3;
	}`
	schemaPath := filepath.Join(tempDir, "person-schema.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","page":1,"result":2.5}`

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	serializationProvider.SetSchemaIDSerializer(serde.HeaderSchemaIDSerializer)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "PROTOBUF",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	headers, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema("topic1-value", tempDir, serde.ValueSerde, &kafka.Message{
		Value:   data,
		Headers: headers,
	})
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", headers, data)
	req.Nil(err)
	req.JSONEq(expectedString, actualString)
}

func TestProtobufSerdesReference(t *testing.T) {
	req := require.New(t)

	tempDir, err := os.MkdirTemp(tempDir, "protobuf")
	req.NoError(err)
	defer os.RemoveAll(tempDir)

	referenceString := `syntax = "proto3";

package test;

message Address {
 string city = 1;
}
`

	// Reference schema should be registered from user side prior to be used as reference
	// So subject and schema version will be known value at this time
	referencePath := filepath.Join(tempDir, "address.proto")
	req.NoError(os.WriteFile(referencePath, []byte(referenceString), 0644))

	schemaString := `syntax = "proto3";

package test;

import "address.proto";

message Person {
 string name = 1;
 test.Address address = 2;
 int32 result = 3;
}
`
	schemaPath := filepath.Join(tempDir, "person.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","address":{"city":"LA"},"result":2}`

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{"address.proto": referencePath})
	req.Nil(err)

	// Explicitly register the reference schema and root schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	referenceInfo := schemaregistry.SchemaInfo{
		Schema:     referenceString,
		SchemaType: "PROTOBUF",
	}
	_, err = client.Register("address.proto", referenceInfo, false)
	req.Nil(err)

	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "PROTOBUF",
		References: []schemaregistry.Reference{
			{
				Name:    "address.proto",
				Subject: "address.proto",
				Version: 1,
			},
		},
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema("topic1-value", tempDir, serde.ValueSerde, &kafka.Message{Value: data})
	req.Nil(err)
	str, err := deserializationProvider.Deserialize("topic1", nil, data)
	req.Nil(err)
	req.JSONEq(str, expectedString)

	// Deserialize again but without the reference file already stored locally
	err = os.Remove(referencePath)
	req.Nil(err)
	err = deserializationProvider.LoadSchema("topic1-value", tempDir, serde.ValueSerde, &kafka.Message{Value: data})
	req.Nil(err)
	str, err = deserializationProvider.Deserialize("topic1", nil, data)
	req.NotNil(err)
	req.JSONEq(str, expectedString)
}

func TestProtobufSerdesInvalid(t *testing.T) {
	req := require.New(t)

	tempDir, err := os.MkdirTemp(tempDir, "protobuf")
	req.NoError(err)
	defer os.RemoveAll(tempDir)

	schemaString := `
	syntax = "proto3";
	message Person {
	  string name = 1;
	  int32 page = 2;
	  int32 result = 3;
	}`
	schemaPath := filepath.Join(tempDir, "person-invalid.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "PROTOBUF",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	exampleString := `{"name":"abc","page":1,"result":2}`
	_, data, err := serializationProvider.Serialize("topic1", exampleString)
	req.Nil(err)

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema("topic1-value", tempDir, serde.ValueSerde, &kafka.Message{Value: data})
	req.Nil(err)

	brokenString := `{"name":"abc`
	brokenBytes := []byte{0, 10, 3, 97, 98, 99, 16}

	_, _, err = serializationProvider.Serialize("topic1", brokenString)
	req.EqualError(err, "the protobuf document is invalid")

	_, err = deserializationProvider.Deserialize("topic1", nil, brokenBytes)
	req.Regexp("^failed to deserialize payload: parsed invalid message index count", err)

	invalidString := `{"page":"abc"}`
	invalidBytes := []byte{0, 12, 3, 97, 98, 99, 16, 1, 24, 2}

	_, _, err = serializationProvider.Serialize("topic1", invalidString)
	req.EqualError(err, "the protobuf document is invalid")

	_, err = deserializationProvider.Deserialize("topic1", nil, invalidBytes)
	req.Regexp("^failed to deserialize payload: parsed invalid message index count", err)
}

func TestProtobufSerdesNestedValid(t *testing.T) {
	req := require.New(t)

	tempDir, err := os.MkdirTemp(tempDir, "protobuf")
	req.NoError(err)
	defer os.RemoveAll(tempDir)

	schemaString := `
	syntax = "proto3";
	message Input {
		string name = 1;
		int32 id = 2;
		Address add = 3;
		PhoneNumber phones = 4;
		message PhoneNumber {
			string number = 1;
		}
		message Address {
			string zip = 1;
			string street = 2;
		}
	}`
	schemaPath := filepath.Join(tempDir, "person-nested.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","id":2,"add":{"zip":"123","street":"def"},"phones":{"number":"234"}}`

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "PROTOBUF",
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema("topic1-value", tempDir, serde.ValueSerde, &kafka.Message{Value: data})
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", nil, data)
	req.Nil(err)
	req.JSONEq(expectedString, actualString)
}

func TestProtobufSerdesValidWithRuleSet(t *testing.T) {
	req := require.New(t)

	tempDir, err := os.MkdirTemp(tempDir, "protobuf")
	req.NoError(err)
	defer os.RemoveAll(tempDir)

	t.Setenv(localKmsSecretMacro, localKmsSecretValueDefault)

	schemaString := `
	syntax = "proto3";

	import "confluent/meta.proto";
	package test;

	message Person {
     string name = 1 [
		(confluent.field_meta).tags = "PII"
     ];
     int32 page = 2;
     double result = 3;
	}`
	schemaPath := filepath.Join(tempDir, "person-schema-ruleset.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","page":1,"result":2.5}`

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", -1, SchemaRegistryAuth{})
	req.Nil(err)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	// CSFLE specific rules during schema registration
	encRule := schemaregistry.Rule{
		Name: "protobuf-encrypt",
		Kind: "TRANSFORM",
		Mode: "WRITEREAD",
		Type: "ENCRYPT",
		Tags: []string{"PII"},
		Params: map[string]string{
			"encrypt.kek.name":   "kek1",
			"encrypt.kms.type":   "local-kms",
			"encrypt.kms.key.id": "mykey",
		},
		OnFailure: "ERROR,NONE",
	}
	ruleSet := schemaregistry.RuleSet{
		DomainRules: []schemaregistry.Rule{encRule},
	}

	// Explicitly register the schema to have a schemaId with mock SR client
	client := serializationProvider.GetSchemaRegistryClient()
	info := schemaregistry.SchemaInfo{
		Schema:     schemaString,
		SchemaType: "PROTOBUF",
		RuleSet:    &ruleSet,
	}
	_, err = client.Register("topic1-value", info, false)
	req.Nil(err)

	_, data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", SchemaRegistryAuth{}, client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema("topic1-value", tempDir, serde.ValueSerde, &kafka.Message{Value: data})
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", nil, data)
	req.Nil(err)
	req.JSONEq(expectedString, actualString)
}

func createTempDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	err := os.MkdirAll(dir, 0755)
	return dir, err
}

func readSchemaReferences(references string) ([]schemaregistry.Reference, error) {
	if references == "" {
		return []schemaregistry.Reference{}, nil
	}

	data, err := os.ReadFile(references)
	if err != nil {
		return nil, err
	}

	var refs []schemaregistry.Reference
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, err
	}

	return refs, nil
}
