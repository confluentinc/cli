package kafka

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"io"
	"io/ioutil"
	"path/filepath"

	configv1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/serdes"
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

func (h *GroupHandler) RequestSchema(value []byte) (string, map[string]string, error) {
	if len(value) < messageOffset {
		return "", nil, errors.New(errors.FailedToFindSchemaIDErrorMsg)
	}

	// Retrieve schema from cluster only if schema is specified.
	schemaID := int32(binary.BigEndian.Uint32(value[1:messageOffset])) // schema id is stored as a part of message meta info

	// Create temporary file to store schema retrieved (also for cache). Retry if get error retriving schema or writing temp schema file
	tempStorePath := filepath.Join(h.Properties.SchemaPath, fmt.Sprintf("%d.txt", schemaID))
	tempRefStorePath := filepath.Join(h.Properties.SchemaPath, fmt.Sprintf("%d.ref", schemaID))
	referencePathMap := map[string]string{}
	var references []srsdk.SchemaReference
	if !fileExists(tempStorePath) || !fileExists(tempRefStorePath) {
		// TODO: add handler for writing schema failure
		schemaString, _, err := h.SrClient.DefaultApi.GetSchema(h.Ctx, schemaID, nil)
		if err != nil {
			return "", nil, err
		}
		err = ioutil.WriteFile(tempStorePath, []byte(schemaString.Schema), 0644)
		if err != nil {
			return "", nil, err
		}

		refBytes, err := json.Marshal(schemaString.References)
		if err != nil {
			return "", nil, err
		}
		err = ioutil.WriteFile(tempRefStorePath, refBytes, 0644)
		if err != nil {
			return "", nil, err
		}
		references = schemaString.References
	} else {
		refBlob, err := ioutil.ReadFile(tempRefStorePath)
		if err != nil {
			return "", nil, err
		}
		err = json.Unmarshal(refBlob, &references)
		if err != nil {
			return "", nil, err
		}
	}

	// Store the references in temporary files
	referencePathMap, err := storeSchemaReferences(references, h.SrClient, h.Ctx)
	if err != nil {
		return "", nil, err
	}

	return tempStorePath, referencePathMap, nil
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
		schemaPath, referencePathMap, err := h.RequestSchema(value)
		if err != nil {
			return err
		}
		// Message body is encoded after 5 bytes of meta information.
		value = value[messageOffset:]
		err = deserializationProvider.LoadSchema(schemaPath, referencePathMap)
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
