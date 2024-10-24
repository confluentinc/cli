package serdes

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
)

type AvroSerializationProvider struct {
	ser      *avrov2.Serializer
	schemaId int
	mode     string
}

func (a *AvroSerializationProvider) InitSerializer(srClientUrl, srClusterId, mode, srApiKey, srApiSecret, token string, schemaId int) error {
	var serdeClientConfig *schemaregistry.Config
	if srClientUrl == mockClientUrl {
		serdeClientConfig = schemaregistry.NewConfig(srClientUrl)
	} else if srApiKey != "" && srApiSecret != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBasicAuthentication(srClientUrl, srApiKey, srApiSecret)
	} else if token != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBearerAuthentication(srClientUrl, token, srClusterId, "")
	} else {
		return fmt.Errorf("schema registry client authentication should be provider to initialize serializer")
	}

	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	// Configure the serde settings
	// If schemaId > 0 then use the intended schema ID
	// otherwise use the latest schema ID
	serdeConfig := avrov2.NewSerializerConfig()
	serdeConfig.AutoRegisterSchemas = false
	serdeConfig.UseLatestVersion = true
	if schemaId > 0 {
		serdeConfig.UseSchemaID = schemaId
		serdeConfig.UseLatestVersion = false
	}

	var serdeType serde.Type
	if mode == "key" {
		serdeType = serde.KeySerde
	} else if mode == "value" {
		serdeType = serde.ValueSerde
	} else {
		return fmt.Errorf("unknown serialization mode: %s", mode)
	}

	ser, err := avrov2.NewSerializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to initialize AVRO serializer: %w", err)
	}

	a.ser = ser
	a.schemaId = schemaId
	if schemaId < 0 {
		a.schemaId = 1
	}
	a.mode = mode
	return nil
}

func (a *AvroSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (a *AvroSerializationProvider) GetSchemaName() string {
	return avroSchemaBackendName
}

func (a *AvroSerializationProvider) Serialize(topic, message string) ([]byte, error) {
	// Step#1: Fetch the schemaInfo based on subject and schema ID
	schemaObj, err := a.GetSchemaRegistryClient().GetBySubjectAndID(topic+"-"+a.mode, a.schemaId)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Step#2: Prepare the Codec based on schemaInfo
	schemaString := schemaObj.Schema
	codec, err := goavro.NewCodec(schemaString)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Step#3: Convert the Avro message data in JSON text format into Go native
	// data types in accordance with the Avro schema supplied when creating the Codec
	object, _, err := codec.NativeFromTextual([]byte(message))
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Step#4: Serialize the Go native data object with the confluent-kafka-go Serialize() library
	payload, err := a.ser.Serialize(topic, &object)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return payload, nil
}

// GetSchemaRegistryClient This getter function is used in mock testing
// as serializer and deserializer have to share the same SR client instance
func (a *AvroSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return a.ser.Client
}
