package kafka

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	ckafka "github.com/confluentinc/confluent-kafka-go-dev/kafka"
)



// sadfsdf
func NewProducer(kafka *config.KafkaClusterConfig, clientID string) (*ckafka.Producer, error) {
	configMap := getProducerConfigMap(kafka, clientID)
	return ckafka.NewProducer(configMap)
}


func getProducerConfigMap(kafka *config.KafkaClusterConfig, clientID string) *ckafka.ConfigMap {
	configMap := &ckafka.ConfigMap{
		"client.id":                             clientID,
		"bootstrap.servers":                     kafka.Bootstrap,
		"security.protocol":                     "SASL_SSL",
		"sasl.username":                         kafka.APIKey,
		"sasl.password":                         kafka.APIKeys[kafka.APIKey].Secret,
		"sasl.mechanism":                        "PLAIN",
		"retry.backoff.ms":                      "250",
		"request.timeout.ms":                    "10000",
		"ssl.endpoint.identification.algorithm": "https",
	}
	return configMap
}

// NewSaramaConsumer returns a sarama.ConsumerGroup configured for the CLI config
func NewConsumer(group string, kafka *config.KafkaClusterConfig, clientID string, beginning bool) (*ckafka.Consumer, error) {
	configMap := getConsumerConfigMap(group, kafka, clientID, beginning)
	return ckafka.NewConsumer(configMap)
}


func getConsumerConfigMap(group string, kafka *config.KafkaClusterConfig, clientID string, beginning bool) *ckafka.ConfigMap {
	var autoOffsetReset string
	if beginning {
		autoOffsetReset = "earliest"
	} else {
		autoOffsetReset = "latest"
	}
	configMap := &ckafka.ConfigMap{
		"group.id":                              group,
		"client.id":                             clientID,
		"bootstrap.servers":                     kafka.Bootstrap,
		"security.protocol":                     "SASL_SSL",
		"sasl.username":                         kafka.APIKey,
		"sasl.password":                         kafka.APIKeys[kafka.APIKey].Secret,
		"sasl.mechanism":                        "PLAIN",
		"ssl.endpoint.identification.algorithm": "https",
		"auto.offset.reset":                     autoOffsetReset,
	}
	return configMap
}
