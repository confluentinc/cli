package serdes

import "errors"

func GetSerializationProvider(valueFormat string) (SerializationProvider, error) {
	var provider SerializationProvider
	// Will add other providers in later commits.
	//if valueFormat == "AVRO" {
	//	provider = new(AvroSerializationProvider)
	//} else if valueFormat == "PROTOBUF" {
	//	provider = new(ProtoSerializationProvider)
	//} else if valueFormat == "JSON" {
	//	provider = new(JsonSerializationProvider)
	//} else if valueFormat == "STRING" {
	if valueFormat == "RAW" {
		provider = new(RawSerializationProvider)
	} else {
		return nil, errors.New("Unknown value format type.")
	}
	return provider, nil
}

func GetDeserializationProvider(valueFormat string) (DeserializationProvider, error) {
	var provider DeserializationProvider
	// Will add other providers in later commits.
	//if valueFormat == "AVRO" {
	//	provider = new(AvroDeserializationProvider)
	//} else if valueFormat == "PROTOBUF" {
	//	provider = new(ProtoDeserializationProvider)
	//} else if valueFormat == "JSON" {
	//	provider = new(JsonDeserializationProvider)
	//} else if valueFormat == "STRING" {
	if valueFormat == "RAW" {
		provider = new(RawDeserializationProvider)
	} else {
		return nil, errors.New("Unknown value format type.")
	}
	return provider, nil
}

type SerializationProvider interface {
	encode(string, string) ([]byte, error)
	GetSchemaName() string
}

func Serialize(provider SerializationProvider, str string, schemaPath string) ([]byte, error) {
	return provider.encode(str, schemaPath)
}

type DeserializationProvider interface {
	decode([]byte, string) (string, error)
	GetSchemaName() string
}

func Deserialize(provider DeserializationProvider, data []byte, schemaPath string) (string, error) {
	return provider.decode(data, schemaPath)
}
