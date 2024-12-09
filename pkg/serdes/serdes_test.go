package serdes

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
)

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
	data, err := serializationProvider.Serialize("", "someString")
	req.Nil(err)
	result := bytes.Compare(data, expectedBytes)
	req.Zero(result)

	deserializationProvider, _ := GetDeserializationProvider(stringSchemaName)
	data = []byte{115, 111, 109, 101, 83, 116, 114, 105, 110, 103}
	str, err := deserializationProvider.Deserialize("", data)
	req.Nil(err)
	req.Equal(str, "someString")
}

func TestAvroSerdesValid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"record","name":"myRecord","fields":[{"name":"f1","type":"int"}]}`
	schemaPath := filepath.Join(dir, "avro-schema.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// expectedBytes[0] is the magic byte, expectedBytes[1:5] is the schemaId in BigIndian
	expectedString := `{"f1":123}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 246, 1}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(avroSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", data)
	req.Nil(err)

	req.Equal(expectedString, actualString)
	req.NoError(os.RemoveAll(dir))
}

func TestAvroSerdesInvalid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"record","name":"myRecord","fields":[{"name":"f1","type":"string"}]}`
	schemaPath := filepath.Join(dir, "avro-schema-invalid.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)

	brokenString := `{"f1"`
	brokenBytes := []byte{0, 0, 0, 0, 1, 6, 97}

	_, err = serializationProvider.Serialize("topic1", brokenString)
	req.Regexp(`cannot decode textual record "myRecord": short buffer`, err)

	_, err = deserializationProvider.Deserialize("topic1", brokenBytes)
	req.Regexp("unexpected EOF$", err)

	invalidString := `{"f2": "abc"}`
	_, err = serializationProvider.Serialize("topic1", invalidString)
	req.Regexp(`cannot decode textual map: cannot determine codec: "f2"$`, err)

	req.NoError(os.RemoveAll(dir))
}

func TestAvroSerdesNestedValid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"record","name":"myRecord","fields":[{"name":"f1","type":"string"},{"name":"nestedField","type":{"type":"record","name":"AnotherNestedRecord","fields":[{"name":"nested_field1","type":"float"},{"name":"nested_field2","type":"string"}]}}]}`
	schemaPath := filepath.Join(dir, "avro-schema-nested.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// expectedBytes[0] is the magic byte, expectedBytes[1:5] is the schemaId in BigIndian
	expectedString := `{"f1":"asd","nestedField":{"nested_field1":123.01,"nested_field2":"example"}}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 6, 97, 115, 100, 31, 5, 246, 66, 14, 101, 120, 97, 109, 112, 108, 101}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(avroSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", expectedBytes)
	req.Nil(err)

	req.Equal(expectedString, actualString)
	req.NoError(os.RemoveAll(dir))
}

func TestJsonSerdesValid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string"}},"required":["f1"]}`
	schemaPath := filepath.Join(dir, "json-schema.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 125}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)

	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", expectedBytes)
	req.Nil(err)
	req.Equal(expectedString, actualString)

	req.NoError(os.RemoveAll(dir))
}

func TestJsonSerdesReference(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	// Reference schema should be registered from user side prior to be used as reference
	// So subject and schema version will be known value at this time
	referenceContent := `[{"name":"RefSchema","subject":"topic2-value","version":1}]`
	referenceString := `{"type": "string"}`
	referencePath := filepath.Join(dir, "json-reference.json")
	req.NoError(os.WriteFile(referencePath, []byte(referenceContent), 0644))

	// Prepare main schema information
	schemaString := `{"type":"object","properties":{"f1":{"$ref":"RefSchema"}},"required":["f1"]}`
	schemaPath := filepath.Join(dir, "json-schema.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// Read references from local reference schema file
	references, err := readSchemaReferences(referencePath)
	req.Nil(err)

	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{0, 0, 0, 0, 2, 123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 125}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)

	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", expectedBytes)
	req.Nil(err)
	req.Equal(expectedString, actualString)

	req.NoError(os.RemoveAll(dir))
}

func TestJsonSerdesInvalid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string"}},"required":["f1"]}`
	schemaPath := filepath.Join(dir, "json-schema-invalid.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)

	brokenString := `{"f1":`
	brokenBytes := []byte{123, 34, 102, 50}

	_, err = serializationProvider.Serialize("topic1", brokenString)
	req.Regexp("unexpected end of JSON input$", err)

	_, err = deserializationProvider.Deserialize("topic1", brokenBytes)
	req.Regexp("unknown magic byte$", err)

	invalidString := `{"f2": "abc"}`
	invalidBytes := []byte{123, 34, 102, 50, 34, 58, 34, 97, 115, 100, 34, 125}

	_, err = serializationProvider.Serialize("topic1", invalidString)
	req.Regexp("missing properties: 'f1'$", err)

	_, err = deserializationProvider.Deserialize("topic1", invalidBytes)
	req.Regexp("unknown magic byte$", err)

	req.NoError(os.RemoveAll(dir))
}

func TestJsonSerdesNestedValid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"object","name":"myRecord","properties":{"f1":{"type":"string"},"nestedField":{"type":"object","name":"AnotherNestedRecord","properties":{"nested_field1":{"type":"number"},"nested_field2":{"type":"string"}}}},"required":["f1","nestedField"]}`
	schemaPath := filepath.Join(dir, "json-schema-nested.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// expectedBytes[0] is the magic byte, expectedBytes[1:5] is the schemaId in BigIndian
	expectedString := `{"f1":"asd","nestedField":{"nested_field1":123.01,"nested_field2":"example"}}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 44, 34, 110, 101, 115, 116, 101, 100, 70, 105, 101, 108, 100, 34,
		58, 123, 34, 110, 101, 115, 116, 101, 100, 95, 102, 105, 101, 108, 100, 49, 34, 58, 49, 50, 51, 46, 48, 49, 44, 34, 110, 101, 115, 116, 101, 100, 95,
		102, 105, 101, 108, 100, 50, 34, 58, 34, 101, 120, 97, 109, 112, 108, 101, 34, 125, 125}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	// Initialize the mock deserializer
	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)

	actualString, err := deserializationProvider.Deserialize("topic1", expectedBytes)
	req.Nil(err)

	req.Equal(expectedString, actualString)
	req.NoError(os.RemoveAll(dir))
}

func TestProtobufSerdesValid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `
	syntax = "proto3";
	message Person {
	  string name = 1;
	  int32 page = 2;
	  double result = 3;
	}`
	schemaPath := filepath.Join(dir, "person-schema.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","page":1,"result":2.5}`

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", data)
	req.Nil(err)
	req.JSONEq(expectedString, actualString)

	req.NoError(os.RemoveAll(dir))
}

func TestProtobufSerdesReference(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	referenceString := `syntax = "proto3";

package io.confluent;

message Address {
  string city = 1;
}
`

	// Reference schema should be registered from user side prior to be used as reference
	// So subject and schema version will be known value at this time
	referencePath := filepath.Join(dir, "address.proto")
	req.NoError(os.WriteFile(referencePath, []byte(referenceString), 0644))

	schemaString := `syntax = "proto3";

package io.confluent;

import "address.proto";

message Person {
  string name = 1;
  io.confluent.Address address = 2;
  int32 result = 3;
}
`
	schemaPath := filepath.Join(dir, "person.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","address":{"city":"LA"},"result":2}`

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{"address.proto": referencePath})
	req.Nil(err)
	str, err := deserializationProvider.Deserialize("topic1", data)
	req.Nil(err)
	req.JSONEq(str, expectedString)

	req.NoError(os.RemoveAll(dir))
}

func TestProtobufSerdesInvalid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `
	syntax = "proto3";
	message Person {
	  string name = 1;
	  int32 page = 2;
	  int32 result = 3;
	}`
	schemaPath := filepath.Join(dir, "person-invalid.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	brokenString := `{"name":"abc`
	brokenBytes := []byte{0, 10, 3, 97, 98, 99, 16}

	_, err = serializationProvider.Serialize("topic1", brokenString)
	req.EqualError(err, "the protobuf document is invalid")

	_, err = deserializationProvider.Deserialize("topic1", brokenBytes)
	req.Regexp("^failed to deserialize payload:.*Subject Not Found$", err)

	invalidString := `{"page":"abc"}`
	invalidBytes := []byte{0, 12, 3, 97, 98, 99, 16, 1, 24, 2}

	_, err = serializationProvider.Serialize("topic1", invalidString)
	req.EqualError(err, "the protobuf document is invalid")

	_, err = deserializationProvider.Deserialize("topic1", invalidBytes)
	req.Regexp("^failed to deserialize payload:.*Subject Not Found$", err)

	req.NoError(os.RemoveAll(dir))
}

func TestProtobufSerdesNestedValid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

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
	schemaPath := filepath.Join(dir, "person-nested.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","id":2,"add":{"zip":"123","street":"def"},"phones":{"number":"234"}}`

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer(mockClientUrl, "", "value", "", "", "", -1)
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

	data, err := serializationProvider.Serialize("topic1", expectedString)
	req.Nil(err)

	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	err = deserializationProvider.InitDeserializer(mockClientUrl, "", "value", "", "", "", client)
	req.Nil(err)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	actualString, err := deserializationProvider.Deserialize("topic1", data)
	req.Nil(err)
	req.JSONEq(expectedString, actualString)

	req.NoError(os.RemoveAll(dir))
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
