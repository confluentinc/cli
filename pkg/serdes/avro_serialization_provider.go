package serdes

import (
	"fmt"
	"os"

	"github.com/linkedin/goavro/v2"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/cel"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/awskms"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/azurekms"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/gcpkms"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/hcvault"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/encryption/localkms"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/rules/jsonata"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
)

type AvroSerializationProvider struct {
	ser      *avrov2.Serializer
	schemaId int
	mode     string
}

func (a *AvroSerializationProvider) InitSerializer(srClientUrl, srClusterId, mode string, schemaId int, srAuth SchemaRegistryAuth) error {
	var serdeClientConfig *schemaregistry.Config
	if srClientUrl == mockClientUrl {
		serdeClientConfig = schemaregistry.NewConfig(srClientUrl)
	} else if srAuth.ApiKey != "" && srAuth.ApiSecret != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBasicAuthentication(srClientUrl, srAuth.ApiKey, srAuth.ApiSecret)
	} else if srAuth.Token != "" {
		serdeClientConfig = schemaregistry.NewConfigWithBearerAuthentication(srClientUrl, srAuth.Token, srClusterId, "")
	} else {
		return fmt.Errorf("schema registry client authentication should be provider to initialize serializer")
	}
	serdeClientConfig.SslCaLocation = srAuth.CertificateAuthorityPath
	serdeClientConfig.SslCertificateLocation = srAuth.ClientCertPath
	serdeClientConfig.SslKeyLocation = srAuth.ClientKeyPath
	serdeClient, err := schemaregistry.NewClient(serdeClientConfig)

	// Register the KMS drivers and the field-level encryption executor
	awskms.Register()
	azurekms.Register()
	gcpkms.Register()
	hcvault.Register()
	localkms.Register()
	encryption.Register()
	cel.Register()
	jsonata.Register()

	if err != nil {
		return fmt.Errorf("failed to create serializer-specific Schema Registry client: %w", err)
	}

	// Configure the serde settings
	// If schemaId > 0 then use the intended schema ID
	// otherwise use the latest schema ID
	serdeConfig := avrov2.NewSerializerConfig()
	serdeConfig.AutoRegisterSchemas = false
	serdeConfig.UseLatestVersion = true

	// local KMS secret is only set and used during local testing with ruleSet
	if localKmsSecretValue := os.Getenv(localKmsSecretMacro); srClientUrl == mockClientUrl && localKmsSecretValue != "" {
		serdeConfig.RuleConfig = map[string]string{
			localKmsSecretKey: localKmsSecretValue,
		}
	}

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

func (a *AvroSerializationProvider) Serialize(topic, message string) ([]kafka.Header, []byte, error) {
	// Step#1: Fetch the schemaInfo based on subject and schema ID
	schemaObj, err := a.GetSchemaRegistryClient().GetBySubjectAndID(topic+"-"+a.mode, a.schemaId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Step#2: Prepare the Codec based on schemaInfo
	schemaString := schemaObj.Schema
	codec, err := goavro.NewCodec(schemaString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Step#3: Convert the Avro message data in JSON text format into Go native
	// data types in accordance with the Avro schema supplied when creating the Codec
	object, _, err := codec.NativeFromTextual([]byte(message))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize message: %w", err)
	}

	// Step#4: Fetch the Go native data object, cast it into generic map for Serialize()
	// Note: the suggested argument to pass to Serialize() library function is:
	// - pointer to a generic map consistent with the schema during registration
	// - a materialized object consistent with the schema during registration
	// Passing the Go native object directly could cause issues during ruleSet execution
	v, ok := object.(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("failed to serialize message: unexpected message type assertion result")
	}
	headers, payload, err := a.ser.SerializeWithHeaders(topic, &v)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return headers, payload, nil
}

// GetSchemaRegistryClient This getter function is used in mock testing
// as serializer and deserializer have to share the same SR client instance
func (a *AvroSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return a.ser.Client
}

// For unit testing purposes
func (a *AvroSerializationProvider) SetSchemaIDSerializer(headerSerializer serde.SchemaIDSerializerFunc) {
	a.ser.SchemaIDSerializer = headerSerializer
}
