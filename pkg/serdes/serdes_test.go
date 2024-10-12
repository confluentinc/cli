package serdes

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/stretchr/testify/require"
)

func TestGetSerializationProvider(t *testing.T) {
	req := require.New(t)
	valueFormats := []string{avroSchemaName, jsonSchemaName, protobufSchemaName}
	schemaNames := []string{avroSchemaBackendName, jsonSchemaBackendName, protobufSchemaBackendName}

	for idx, valueFormat := range valueFormats {
		provider, err := GetSerializationProvider(valueFormat)
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
	expectedBytes := []byte{115, 111, 109, 101, 115, 116, 114, 105, 110, 103}
	data, err := serializationProvider.Serialize("", "somestring")
	req.Nil(err)
	result := bytes.Compare(data, expectedBytes)
	req.Zero(result)

	deserializationProvider, _ := GetDeserializationProvider(stringSchemaName)
	data = []byte{115, 111, 109, 101, 115, 116, 114, 105, 110, 103}
	str, err := deserializationProvider.Deserialize("", data)
	req.Nil(err)
	req.Equal(str, "somestring")
}

func TestAvroSerdesValid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"record","name":"myrecord","fields":[{"name":"f1","type":"string"}]}`
	schemaPath := filepath.Join(dir, "avro-schema.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// expectedBytes[0] is the magic byte, expectedBytes[1:5] is the schemaId in BigIndian
	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 6, 97, 115, 100}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err = serializationProvider.InitSerializer("mock://", "value", -1)
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
	err = deserializationProvider.InitDeserializer("mock://", "value", client)
	req.Nil(err)

	str, err := deserializationProvider.Deserialize("topic1", expectedBytes)
	req.Nil(err)

	req.Equal(str, expectedString)
	req.NoError(os.RemoveAll(dir))
}

func TestAvroSerdesInvalid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"record","name":"myrecord","fields":[{"name":"f1","type":"string"}]}`
	schemaPath := filepath.Join(dir, "avro-schema.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(avroSchemaName)
	err = serializationProvider.InitSerializer("mock://", "value", -1)
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

	err = deserializationProvider.InitDeserializer("mock://", "value", client)
	req.Nil(err)

	brokenString := `{"f1"`
	brokenBytes := []byte{0, 0, 0, 0, 1, 6, 97}

	_, err = serializationProvider.Serialize("topic1", brokenString)
	req.Regexp("unexpected end of JSON input$", err)

	_, err = deserializationProvider.Deserialize("topic1", brokenBytes)
	req.Regexp("unexpected EOF$", err)

	invalidString := `{"f2": "abc"}`
	_, err = serializationProvider.Serialize("topic1", invalidString)
	req.Regexp("missing required field f1$", err)

	req.NoError(os.RemoveAll(dir))
}

func TestJsonSerdesValid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string"}},"required":["f1"]}`
	schemaPath := filepath.Join(dir, "json-demo.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{0, 0, 0, 0, 1, 123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 125}

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err = serializationProvider.InitSerializer("mock://", "value", -1)
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
	err = deserializationProvider.InitDeserializer("mock://", "value", client)
	req.Nil(err)

	//err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	//req.Nil(err)
	str, err := deserializationProvider.Deserialize("topic1", expectedBytes)
	req.Nil(err)
	req.Equal(str, expectedString)

	req.NoError(os.RemoveAll(dir))
}

//func TestJsonSerdesReference(t *testing.T) {
//	req := require.New(t)
//
//	dir, err := createTempDir()
//	req.Nil(err)
//
//	referenceString := `{"type": "string"}`
//	referencePath := filepath.Join(dir, "json-reference.json")
//	req.NoError(os.WriteFile(referencePath, []byte(referenceString), 0644))
//
//	schemaString := `{"type":"object","properties":{"f1":{"$ref":"json-reference.json"}},"required":["f1"]}`
//	schemaPath := filepath.Join(dir, "json-demo.json")
//	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))
//
//	expectedString := `{"f1":"asd"}`
//	expectedBytes := []byte{123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 125}
//
//	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
//	err = serializationProvider.LoadSchema(schemaPath, map[string]string{"json-reference.json": referencePath})
//	req.Nil(err)
//	data, err := serializationProvider.Serialize(expectedString)
//	req.Nil(err)
//
//	result := bytes.Compare(expectedBytes, data)
//	req.Zero(result)
//
//	data = expectedBytes
//	deserializationProvider, _ := GetDeserializationProvider(jsonSchemaName)
//	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{"json-reference.json": referencePath})
//	req.Nil(err)
//	str, err := deserializationProvider.Deserialize(data)
//	req.Nil(err)
//	req.Equal(str, expectedString)
//
//	req.NoError(os.RemoveAll(dir))
//}

func TestJsonSerdesInvalid(t *testing.T) {
	req := require.New(t)

	dir, err := createTempDir()
	req.Nil(err)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string"}},"required":["f1"]}`
	schemaPath := filepath.Join(dir, "json-demo.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	// Initialize the mock serializer and use latest schemaId
	serializationProvider, _ := GetSerializationProvider(jsonSchemaName)
	err = serializationProvider.InitSerializer("mock://", "value", -1)
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
	err = deserializationProvider.InitDeserializer("mock://", "value", client)
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

func TestProtobufSerdesValid(t *testing.T) {
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
	schemaPath := filepath.Join(dir, "person.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","page":1,"result":2}`
	expectedBytes := []byte{0, 10, 3, 97, 98, 99, 16, 1, 24, 2}

	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
	err = serializationProvider.InitSerializer("mock://", "value", -1)
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

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	data = expectedBytes
	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
	//err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	//req.Nil(err)
	err = deserializationProvider.InitDeserializer("mock://", "value", client)
	req.Nil(err)
	str, err := deserializationProvider.Deserialize("topic1", data)
	req.Nil(err)
	req.Equal(str, expectedString)

	req.NoError(os.RemoveAll(dir))
}

//func TestProtobufSerdesReference(t *testing.T) {
//	req := require.New(t)
//
//	dir, err := createTempDir()
//	req.Nil(err)
//
//	referenceString := `
//	syntax = "proto3";
//    package io.confluent;
//	message Address {
//	  string city = 1;
//	}`
//
//	referencePath := filepath.Join(dir, "address.proto")
//	req.NoError(os.WriteFile(referencePath, []byte(referenceString), 0644))
//
//	schemaString := `
//    syntax = "proto3";
//    package io.confluent;
//    import "address.proto";
//	message Person {
//	  string name = 1;
//	  io.confluent.Address address = 2;
//	  int32 result = 3;
//	}`
//	schemaPath := filepath.Join(dir, "person.proto")
//	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))
//
//	expectedString := `{"name":"abc","address":{"city":"LA"},"result":2}`
//	expectedBytes := []byte{0, 10, 3, 97, 98, 99, 18, 4, 10, 2, 76, 65, 24, 2}
//
//	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
//	err = serializationProvider.LoadSchema(schemaPath, map[string]string{"address.proto": referencePath})
//	req.Nil(err)
//	data, err := serializationProvider.Serialize(expectedString)
//	req.Nil(err)
//
//	result := bytes.Compare(expectedBytes, data)
//	req.Equal(expectedBytes, data)
//	req.Zero(result)
//
//	data = expectedBytes
//	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
//	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{"address.proto": referencePath})
//	req.Nil(err)
//	str, err := deserializationProvider.Deserialize(data)
//	req.Nil(err)
//	req.Equal(str, expectedString)
//
//	req.NoError(os.RemoveAll(dir))
//}
//
//func TestProtobufSerdesInvalid(t *testing.T) {
//	req := require.New(t)
//
//	dir, err := createTempDir()
//	req.Nil(err)
//
//	schemaString := `
//	syntax = "proto3";
//	message Person {
//	  string name = 1;
//	  int32 page = 2;
//	  int32 result = 3;
//	}`
//	schemaPath := filepath.Join(dir, "person.proto")
//	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))
//
//	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
//	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
//	req.Nil(err)
//	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
//	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
//	req.Nil(err)
//
//	brokenString := `{"name":"abc`
//	brokenBytes := []byte{0, 10, 3, 97, 98, 99, 16}
//
//	_, err = serializationProvider.Serialize(brokenString)
//	req.EqualError(err, "the protobuf document is invalid")
//
//	data := brokenBytes
//	_, err = deserializationProvider.Deserialize(data)
//	req.EqualError(err, "the protobuf document is invalid")
//
//	invalidString := `{"page":"abc"}`
//	invalidBytes := []byte{0, 12, 3, 97, 98, 99, 16, 1, 24, 2}
//
//	_, err = serializationProvider.Serialize(invalidString)
//	req.Error(err)
//
//	data = invalidBytes
//	_, err = deserializationProvider.Deserialize(data)
//	req.Error(err)
//
//	req.NoError(os.RemoveAll(dir))
//}
//
//func TestProtobufSerdesNestedValid(t *testing.T) {
//	req := require.New(t)
//
//	dir, err := createTempDir()
//	req.Nil(err)
//
//	schemaString := `
//	syntax = "proto3";
//	import "google/protobuf/descriptor.proto";
//	message Input {
//		string name = 1;
//		int32 id = 2;  // Unique ID number for this person.
//		Address add = 3;
//		PhoneNumber phones = 4;  //List
//		message PhoneNumber {
//			string number = 1;
//		}
//		message Address {
//			string zip = 1;
//			string street = 2;
//		}
//	}`
//	schemaPath := filepath.Join(dir, "person.proto")
//	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))
//
//	expectedString := `{"name":"abc","id":2,"add":{"zip":"123","street":"def"},"phones":{"number":"234"}}`
//	expectedBytes := []byte{
//		0, 10, 3, 97, 98, 99, 16, 2, 26, 10, 10, 3,
//		49, 50, 51, 18, 3, 100, 101, 102, 34, 5, 10, 3, 50, 51, 52,
//	}
//
//	serializationProvider, _ := GetSerializationProvider(protobufSchemaName)
//	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
//	req.Nil(err)
//	data, err := serializationProvider.Serialize(expectedString)
//	req.Nil(err)
//
//	result := bytes.Compare(expectedBytes, data)
//	req.Zero(result)
//
//	data = expectedBytes
//
//	deserializationProvider, _ := GetDeserializationProvider(protobufSchemaName)
//	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
//	req.Nil(err)
//	str, err := deserializationProvider.Deserialize(data)
//	req.Nil(err)
//	req.Equal(str, expectedString)
//
//	req.NoError(os.RemoveAll(dir))
//}

func createTempDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "ccloud-schema")
	err := os.MkdirAll(dir, 0755)
	return dir, err
}
