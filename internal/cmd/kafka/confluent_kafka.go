package kafka

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"

	configv1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	serdes "github.com/confluentinc/cli/internal/pkg/serdes"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
)

// Schema ID is stored at the [1:5] bytes of a message as meta info (when valid)
const messageOffset = 5

type ConsumerProperties struct {
	PrintKey   bool
	Delimiter  string
	SchemaPath string
}

// GroupHandler instances are used to handle individual topic-partition claims.
type GroupHandler struct {
	SrClient   *srsdk.APIClient
	Ctx        context.Context
	Format     string
	Out        io.Writer
	Properties ConsumerProperties
}

func NewProducer(kafka *configv1.KafkaClusterConfig, clientID string) (*ckafka.Producer, error) {
	configMap, err := getProducerConfigMap(kafka, clientID)
	if err != nil {
		return nil, err
	}
	return ckafka.NewProducer(configMap)
}

// NewConsumer returns a ConsumerGroup configured for the CLI config
func NewConsumer(group string, kafka *configv1.KafkaClusterConfig, clientID string, beginning bool) (*ckafka.Consumer, error) {
	configMap, err := getConsumerConfigMap(group, kafka, clientID, beginning)
	if err != nil {
		return nil, err
	}
	return ckafka.NewConsumer(configMap)
}

func getCommonConfig(kafka *configv1.KafkaClusterConfig, clientID string) *ckafka.ConfigMap {
	return &ckafka.ConfigMap{
		"security.protocol":                     "SASL_SSL",
		"sasl.mechanism":                        "PLAIN",
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"bootstrap.servers":                     kafka.Bootstrap,
		"sasl.username":                         kafka.APIKey,
		"sasl.password":                         kafka.APIKeys[kafka.APIKey].Secret,
	}
}

func GetOnPremProducerCommonConfig(clientID, bootstrap string, enableSSLVerification bool) *ckafka.ConfigMap {
	return &ckafka.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"bootstrap.servers":                     bootstrap,
		"enable.ssl.certificate.verification":   enableSSLVerification,
		"retry.backoff.ms":                      "250",
		"request.timeout.ms":                    "10000",
		// "schema.registry.basic.auth.user.info": "<sr-api-key>:<sr-api-secret>", // TODO
		// "schema.registry.url":                  "<schema-registry-url>",
	}
}

func GetOnPremConsumerCommonConfig(clientID, bootstrap, group string, beginning, enableSSLVerification bool) (*ckafka.ConfigMap, error) {
	configMap := &ckafka.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"group.id":                              group,
		"bootstrap.servers":                     bootstrap,
		"enable.ssl.certificate.verification":   enableSSLVerification,
		// "schema.registry.basic.auth.user.info": "<sr-api-key>:<sr-api-secret>", // TODO
		// "schema.registry.url":                  "<schema-registry-url>",
	}
	var autoOffsetReset string
	if beginning {
		autoOffsetReset = "earliest"
	} else {
		autoOffsetReset = "latest"
	}
	if err := configMap.SetKey("auto.offset.reset", autoOffsetReset); err != nil {
		return nil, err
	}
	return configMap, nil
}

func SetSSLConfig(configMap *ckafka.ConfigMap, caLocation, certLocation, keyLocation, keyPassword string) (*ckafka.ConfigMap, error) {
	if err := configMap.SetKey("security.protocol", "SSL"); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("ssl.ca.location", caLocation); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("ssl.certificate.location", certLocation); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("ssl.key.location", keyLocation); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("ssl.key.password", keyPassword); err != nil {
		return nil, err
	}
	return configMap, nil
}

func SetSASLConfig(configMap *ckafka.ConfigMap, username, password string) (*ckafka.ConfigMap, error) {
	if err := configMap.SetKey("security.protocol", "SASL_SSL"); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("sasl.mechanism", "PLAIN"); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("sasl.username", username); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("sasl.password", password); err != nil {
		return nil, err
	}
	return configMap, nil
}

func getProducerConfigMap(kafka *configv1.KafkaClusterConfig, clientID string) (*ckafka.ConfigMap, error) {
	configMap := getCommonConfig(kafka, clientID)
	if err := configMap.SetKey("retry.backoff.ms", "250"); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("request.timeout.ms", "10000"); err != nil {
		return nil, err
	}
	return configMap, nil
}

func getConsumerConfigMap(group string, kafka *configv1.KafkaClusterConfig, clientID string, beginning bool) (*ckafka.ConfigMap, error) {
	configMap := getCommonConfig(kafka, clientID)
	if err := configMap.SetKey("group.id", group); err != nil {
		return nil, err
	}
	var autoOffsetReset string
	if beginning {
		autoOffsetReset = "earliest"
	} else {
		autoOffsetReset = "latest"
	}
	if err := configMap.SetKey("auto.offset.reset", autoOffsetReset); err != nil {
		return nil, err
	}
	return configMap, nil
}

func (h *GroupHandler) RequestSchema(value []byte) (string, error) {
	if len(value) < messageOffset {
		return "", errors.New(errors.FailedToFindSchemaIDErrorMsg)
	}

	// Retrieve schema from cluster only if schema is specified.
	schemaID := int32(binary.BigEndian.Uint32(value[1:messageOffset])) // schema id is stored as a part of message meta info

	// Create temporary file to store schema retrieved (also for cache). Retry if get error retriving schema or writing temp schema file
	tempStorePath := filepath.Join(h.Properties.SchemaPath, strconv.Itoa(int(schemaID))+".txt")
	if !fileExists(tempStorePath) {
		// TODO: add handler for writing schema failure
		schemaString, _, err := h.SrClient.DefaultApi.GetSchema(h.Ctx, schemaID, nil)
		if err != nil {
			return "", err
		}
		err = ioutil.WriteFile(tempStorePath, []byte(schemaString.Schema), 0644)
		if err != nil {
			return "", err
		}
	}
	return tempStorePath, nil
}

func ConsumeMessage(e *ckafka.Message, h *GroupHandler) error {
	value := e.Value
	if h.Properties.PrintKey {
		key := e.Key
		var keyString string
		if len(key) == 0 {
			keyString = "null"
		} else {
			keyString = string(key)
		}
		_, err := fmt.Fprint(h.Out, keyString+h.Properties.Delimiter)
		if err != nil {
			return err
		}
	}

	deserializationProvider, err := serdes.GetDeserializationProvider(h.Format)
	if err != nil {
		return err
	}

	if h.Format != "string" {
		schemaPath, err := h.RequestSchema(value)
		if err != nil {
			return err
		}
		// Message body is encoded after 5 bytes of meta information.
		value = value[messageOffset:]
		err = deserializationProvider.LoadSchema(schemaPath)
		if err != nil {
			return err
		}
	}
	jsonMessage, err := serdes.Deserialize(deserializationProvider, value)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(h.Out, jsonMessage)
	if err != nil {
		return err
	}

	if e.Headers != nil {
		_, err = fmt.Fprintf(h.Out, "%% Headers: %v\n", e.Headers)
		if err != nil {
			return err
		}
	}
	return nil
}
