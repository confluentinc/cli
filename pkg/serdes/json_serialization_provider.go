package serdes

import (
	"encoding/json"
	"fmt"
	"os"

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
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/jsonschema"
)

type JsonSerializationProvider struct {
	ser *jsonschema.Serializer
}

func (j *JsonSerializationProvider) InitSerializer(srClientUrl, srClusterId, mode string, schemaId int, srAuth SchemaRegistryAuth) error {
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
	serdeConfig := jsonschema.NewSerializerConfig()
	serdeConfig.AutoRegisterSchemas = false
	serdeConfig.UseLatestVersion = true
	serdeConfig.EnableValidation = true

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

	ser, err := jsonschema.NewSerializer(serdeClient, serdeType, serdeConfig)

	if err != nil {
		return fmt.Errorf("failed to initialize JSON serializer: %w", err)
	}

	j.ser = ser
	return nil
}

func (j *JsonSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (j *JsonSerializationProvider) GetSchemaName() string {
	return jsonSchemaBackendName
}

func (j *JsonSerializationProvider) Serialize(topic, message string) ([]kafka.Header, []byte, error) {
	// Convert the plain string message from customer type-in in CLI terminal into generic map
	var result map[string]any
	err := json.Unmarshal([]byte(message), &result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert message string into generic map for serialization: %w", err)
	}

	headers, payload, err := j.ser.SerializeWithHeaders(topic, &result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to serialize message: %w", err)
	}
	return headers, payload, nil
}

// GetSchemaRegistryClient This getter function is used in mock testing
// as serializer and deserializer have to share the same SR client instance
func (j *JsonSerializationProvider) GetSchemaRegistryClient() schemaregistry.Client {
	return j.ser.Client
}

// For unit testing purposes
func (j *JsonSerializationProvider) SetSchemaIDSerializer(headerSerializer serde.SchemaIDSerializerFunc) {
	j.ser.SchemaIDSerializer = headerSerializer
}
