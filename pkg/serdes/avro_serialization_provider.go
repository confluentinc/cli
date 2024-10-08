package serdes

import (
	"fmt"

	"encoding/json"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
)

type AvroSerializationProvider struct {
	ser *avrov2.Serializer
}

func (a *AvroSerializationProvider) InitSerializer(srClientUrl, mode string) error {
	serdeClientConfig := schemaregistry.NewConfig(srClientUrl)
	serdeClientConfig.BasicAuthCredentialsSource = "USER_INFO"
	serdeClientConfig.BasicAuthUserInfo = "IZ7RGM7EFPH6TJP4:gi3a/MpHzh8wQZXcWrbQ+emhZZKOoyI1gQo20EaYoa0EFU4TAH69rAk1KXjm0+sN"

	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)
	fmt.Printf("The srClientUrl is %s, mode is %s\n", srClientUrl, mode)

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	serdeConfig := avrov2.NewSerializerConfig()
	//serdeConfig.AutoRegisterSchemas = false
	//serdeConfig.UseLatestVersion = true
	//serdeConfig.UseSchemaID = 100002
	//serdeConfig.NormalizeSchemas = true

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
		return fmt.Errorf("failed to create serializer: %w", err)
	}

	jsonSer, err := json.Marshal(*ser)
	fmt.Printf("The serializer object is %s\n", string(jsonSer))

	a.ser = ser
	return nil
}

func (a *AvroSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (a *AvroSerializationProvider) GetSchemaName() string {
	return avroSchemaBackendName
}

func (a *AvroSerializationProvider) Serialize(topic string, message any) ([]byte, error) {
	// Convert the plain string message from customer into map
	var result map[string]any
	err := json.Unmarshal([]byte(message.(string)), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message string into map struct for serialization: %w", err)
	}

	payload, err := a.ser.Serialize(topic, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return payload, nil
}
