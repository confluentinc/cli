package serdes

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	sr "github.com/confluentinc/cli/internal/cmd/schema-registry"
)

func TestSerializationProvider(t *testing.T) {
	req := require.New(t)
	valueFormats := []string{AVROSCHEMANAME, PROTOBUFSCHEMANAME, JSONSCHEMANAME, RAWSCHEMANAME}
	schemaNames := []string{AVROSCHEMABACKEND, PROTOBUFSCHEMABACKEND, JSONSCHEMABACKEND, RAWSCHEMANAME}

	for idx, valueFormat := range valueFormats {
		provider, err := GetSerializationProvider(valueFormat)
		req.Equal(provider.GetSchemaName(), schemaNames[idx])
		req.Nil(err)
	}

	provider, err := GetSerializationProvider("UNKNOWN")
	req.Nil(provider)
	req.EqualError(err, "unknown value schema format")
}

func TestDeserializationProvider(t *testing.T) {
	req := require.New(t)
	valueFormats := []string{AVROSCHEMANAME, PROTOBUFSCHEMANAME, JSONSCHEMANAME, RAWSCHEMANAME}
	schemaNames := []string{AVROSCHEMABACKEND, PROTOBUFSCHEMABACKEND, JSONSCHEMABACKEND, RAWSCHEMANAME}

	for idx, valueFormat := range valueFormats {
		provider, err := GetDeserializationProvider(valueFormat)
		req.Equal(provider.GetSchemaName(), schemaNames[idx])
		req.Nil(err)
	}

	provider, err := GetDeserializationProvider("UNKNOWN")
	req.Nil(provider)
	req.EqualError(err, "unknown value schema format")
}

func TestStringSerdes(t *testing.T) {
	req := require.New(t)

	serializationProvider, _ := GetSerializationProvider(RAWSCHEMANAME)
	expectedBytes := []byte{115, 111, 109, 101, 115, 116, 114, 105, 110, 103}
	data, err := serializationProvider.encode("somestring")
	req.Nil(err)
	result := bytes.Compare(data, expectedBytes)
	req.Zero(result)

	deserializationProvider, _ := GetDeserializationProvider(RAWSCHEMANAME)
	data = []byte{115, 111, 109, 101, 115, 116, 114, 105, 110, 103}
	str, err := deserializationProvider.decode(data)
	req.Nil(err)
	req.Equal(str, "somestring")
}

func TestAvroSerdesValid(t *testing.T) {
	req := require.New(t)

	dir, err := sr.CreateTempDir()
	req.Nil(err)

	schemaString := `{"type":"record","name":"myrecord","fields":[{"name":"f1","type":"string"}]}`
	schemaPath := filepath.Join(dir, "avro-schema.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{6, 97, 115, 100}

	serializationProvider, _ := GetSerializationProvider(AVROSCHEMANAME)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	data, err := serializationProvider.encode(expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	data = []byte{6, 97, 115, 100}

	deserializationProvider, _ := GetDeserializationProvider(AVROSCHEMANAME)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	str, err := deserializationProvider.decode(data)
	req.Nil(err)
	req.Equal(str, expectedString)

	req.NoError(os.RemoveAll(dir))
}

func TestAvroSerdesInvalid(t *testing.T) {
	req := require.New(t)

	dir, err := sr.CreateTempDir()
	req.Nil(err)

	schemaString := `{"type":"record","name":"myrecord","fields":[{"name":"f1","type":"string"}]}`
	schemaPath := filepath.Join(dir, "avro-schema.txt")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	serializationProvider, _ := GetSerializationProvider(AVROSCHEMANAME)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	deserializationProvider, _ := GetDeserializationProvider(AVROSCHEMANAME)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	brokenString := `{"f1"`
	brokenBytes := []byte{6, 97}

	_, err = serializationProvider.encode(brokenString)
	req.Regexp("^cannot decode textual record", err)

	data := brokenBytes
	_, err = deserializationProvider.decode(data)
	req.Regexp("^cannot decode binary record", err)

	invalidString := `{"f2": "abc"}`
	_, err = serializationProvider.encode(invalidString)
	req.Regexp("^cannot decode textual record", err)

	req.NoError(os.RemoveAll(dir))
}

func TestJsonSerdesValid(t *testing.T) {
	req := require.New(t)

	dir, err := sr.CreateTempDir()
	req.Nil(err)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string"}},"required":["f1"]}`
	schemaPath := filepath.Join(dir, "json-demo.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 125}

	serializationProvider, _ := GetSerializationProvider(JSONSCHEMANAME)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	data, err := serializationProvider.encode(expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	data = expectedBytes
	deserializationProvider, _ := GetDeserializationProvider(JSONSCHEMANAME)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	str, err := deserializationProvider.decode(data)
	req.Nil(err)
	req.Equal(str, expectedString)

	req.NoError(os.RemoveAll(dir))
}

func TestJsonSerdesReference(t *testing.T) {
	req := require.New(t)

	dir, err := sr.CreateTempDir()
	req.Nil(err)

	referenceString := `{"type": "string"}`
	referencePath := filepath.Join(dir, "json-reference.json")
	req.NoError(os.WriteFile(referencePath, []byte(referenceString), 0644))

	schemaString := `{"type":"object","properties":{"f1":{"$ref":"json-reference.json"}},"required":["f1"]}`
	schemaPath := filepath.Join(dir, "json-demo.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"f1":"asd"}`
	expectedBytes := []byte{123, 34, 102, 49, 34, 58, 34, 97, 115, 100, 34, 125}

	serializationProvider, _ := GetSerializationProvider(JSONSCHEMANAME)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{"json-reference.json": referencePath})
	req.Nil(err)
	data, err := serializationProvider.encode(expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	data = expectedBytes
	deserializationProvider, _ := GetDeserializationProvider(JSONSCHEMANAME)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{"json-reference.json": referencePath})
	req.Nil(err)
	str, err := deserializationProvider.decode(data)
	req.Nil(err)
	req.Equal(str, expectedString)

	req.NoError(os.RemoveAll(dir))
}

func TestJsonSerdesInvalid(t *testing.T) {
	req := require.New(t)

	dir, err := sr.CreateTempDir()
	req.Nil(err)

	schemaString := `{"type":"object","properties":{"f1":{"type":"string"}},"required":["f1"]}`
	schemaPath := filepath.Join(dir, "json-demo.json")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	serializationProvider, _ := GetSerializationProvider(JSONSCHEMANAME)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	deserializationProvider, _ := GetDeserializationProvider(JSONSCHEMANAME)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	brokenString := `{"f1":`
	brokenBytes := []byte{123, 34, 102, 50}

	_, err = serializationProvider.encode(brokenString)
	req.EqualError(err, "unexpected EOF")

	data := brokenBytes
	_, err = deserializationProvider.decode(data)
	req.EqualError(err, "unexpected EOF")

	invalidString := `{"f2": "abc"}`
	invalidBytes := []byte{123, 34, 102, 50, 34, 58, 34, 97, 115, 100, 34, 125}

	_, err = serializationProvider.encode(invalidString)
	req.EqualError(err, "the JSON document is invalid")

	data = invalidBytes
	_, err = deserializationProvider.decode(data)
	req.EqualError(err, "the JSON document is invalid")

	req.NoError(os.RemoveAll(dir))
}

func TestProtobufSerdesValid(t *testing.T) {
	req := require.New(t)

	dir, err := sr.CreateTempDir()
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

	serializationProvider, _ := GetSerializationProvider(PROTOBUFSCHEMANAME)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	data, err := serializationProvider.encode(expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	data = expectedBytes
	deserializationProvider, _ := GetDeserializationProvider(PROTOBUFSCHEMANAME)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	str, err := deserializationProvider.decode(data)
	req.Nil(err)
	req.Equal(str, expectedString)

	req.NoError(os.RemoveAll(dir))
}

func TestProtobufSerdesReference(t *testing.T) {
	req := require.New(t)

	dir, err := sr.CreateTempDir()
	req.Nil(err)

	referenceString := `
	syntax = "proto3";
    package io.confluent;
	message Address {
	  string city = 1;
	}`

	referencePath := filepath.Join(dir, "address.proto")
	req.NoError(os.WriteFile(referencePath, []byte(referenceString), 0644))

	schemaString := `
    syntax = "proto3";
    package io.confluent;
    import "address.proto";
	message Person {
	  string name = 1;
	  io.confluent.Address address = 2;
	  int32 result = 3;
	}`
	schemaPath := filepath.Join(dir, "person.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","address":{"city":"LA"},"result":2}`
	expectedBytes := []byte{0, 10, 3, 97, 98, 99, 18, 4, 10, 2, 76, 65, 24, 2}

	serializationProvider, _ := GetSerializationProvider(PROTOBUFSCHEMANAME)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{"address.proto": referencePath})
	req.Nil(err)
	data, err := serializationProvider.encode(expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Equal(expectedBytes, data)
	req.Zero(result)

	data = expectedBytes
	deserializationProvider, _ := GetDeserializationProvider(PROTOBUFSCHEMANAME)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{"address.proto": referencePath})
	req.Nil(err)
	str, err := deserializationProvider.decode(data)
	req.Nil(err)
	req.Equal(str, expectedString)

	req.NoError(os.RemoveAll(dir))
}

func TestProtobufSerdesInvalid(t *testing.T) {
	req := require.New(t)

	dir, err := sr.CreateTempDir()
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

	serializationProvider, _ := GetSerializationProvider(PROTOBUFSCHEMANAME)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	deserializationProvider, _ := GetDeserializationProvider(PROTOBUFSCHEMANAME)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)

	brokenString := `{"name":"abc`
	brokenBytes := []byte{0, 10, 3, 97, 98, 99, 16}

	_, err = serializationProvider.encode(brokenString)
	req.EqualError(err, "the protobuf document is invalid")

	data := brokenBytes
	_, err = deserializationProvider.decode(data)
	req.EqualError(err, "the protobuf document is invalid")

	invalidString := `{"page":"abc"}`
	invalidBytes := []byte{0, 12, 3, 97, 98, 99, 16, 1, 24, 2}

	_, err = serializationProvider.encode(invalidString)
	req.Error(err)

	data = invalidBytes
	_, err = deserializationProvider.decode(data)
	req.Error(err)

	req.NoError(os.RemoveAll(dir))
}

func TestProtobufSerdesNestedValid(t *testing.T) {
	req := require.New(t)

	dir, err := sr.CreateTempDir()
	req.Nil(err)

	schemaString := `
	syntax = "proto3";
	import "google/protobuf/descriptor.proto";
	message Input {
		string name = 1;
		int32 id = 2;  // Unique ID number for this person.
		Address add = 3;
		PhoneNumber phones = 4;  //List
		message PhoneNumber {
			string number = 1;
		}
		message Address {
			string zip = 1;
			string street = 2;
		}
	}`
	schemaPath := filepath.Join(dir, "person.proto")
	req.NoError(os.WriteFile(schemaPath, []byte(schemaString), 0644))

	expectedString := `{"name":"abc","id":2,"add":{"zip":"123","street":"def"},"phones":{"number":"234"}}`
	expectedBytes := []byte{0, 10, 3, 97, 98, 99, 16, 2, 26, 10, 10, 3,
		49, 50, 51, 18, 3, 100, 101, 102, 34, 5, 10, 3, 50, 51, 52}

	serializationProvider, _ := GetSerializationProvider(PROTOBUFSCHEMANAME)
	err = serializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	data, err := serializationProvider.encode(expectedString)
	req.Nil(err)

	result := bytes.Compare(expectedBytes, data)
	req.Zero(result)

	data = expectedBytes

	deserializationProvider, _ := GetDeserializationProvider(PROTOBUFSCHEMANAME)
	err = deserializationProvider.LoadSchema(schemaPath, map[string]string{})
	req.Nil(err)
	str, err := deserializationProvider.decode(data)
	req.Nil(err)
	req.Equal(str, expectedString)

	req.NoError(os.RemoveAll(dir))
}
