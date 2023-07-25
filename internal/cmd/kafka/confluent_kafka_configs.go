package kafka

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"

	configv1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/properties"
)

type kafkaClientConfigs struct {
	configurations map[string]string
}

type PartitionFilter struct {
	Changed bool
	Index   int32
}

func getCommonConfig(kafka *configv1.KafkaClusterConfig, clientId string) (*ckafka.ConfigMap, error) {
	if err := kafka.DecryptAPIKeys(); err != nil {
		return nil, err
	}

	configMap := &ckafka.ConfigMap{
		"security.protocol":                     "SASL_SSL",
		"sasl.mechanism":                        "PLAIN",
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientId,
		"bootstrap.servers":                     kafka.Bootstrap,
		"sasl.username":                         kafka.APIKey,
		"sasl.password":                         kafka.GetApiSecret(),
	}

	return configMap, nil
}

func getProducerConfigMap(kafka *configv1.KafkaClusterConfig, clientID string) (*ckafka.ConfigMap, error) {
	configMap, err := getCommonConfig(kafka, clientID)
	if err != nil {
		return nil, err
	}
	if err := configMap.SetKey("retry.backoff.ms", "250"); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("request.timeout.ms", "10000"); err != nil {
		return nil, err
	}
	if err := SetProducerDebugOption(configMap); err != nil {
		return nil, err
	}
	return configMap, nil
}

func getConsumerConfigMap(group string, kafka *configv1.KafkaClusterConfig, clientID string) (*ckafka.ConfigMap, error) {
	configMap, err := getCommonConfig(kafka, clientID)
	if err != nil {
		return nil, err
	}
	if err := configMap.SetKey("group.id", group); err != nil {
		return nil, err
	}
	log.CliLogger.Debugf("Created consumer group: %s", group)

	// see explanation: https://www.confluent.io/blog/incremental-cooperative-rebalancing-in-kafka/
	if err := configMap.SetKey("partition.assignment.strategy", "cooperative-sticky"); err != nil {
		return nil, err
	}
	if err := SetConsumerDebugOption(configMap); err != nil {
		return nil, err
	}
	return configMap, nil
}

func getOnPremCommonConfig(clientID, bootstrap, caLocation string) *ckafka.ConfigMap {
	return &ckafka.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"bootstrap.servers":                     bootstrap,
		"enable.ssl.certificate.verification":   true,
		"ssl.ca.location":                       caLocation,
	}
}

func getOnPremProducerConfigMap(cmd *cobra.Command, clientID string) (*ckafka.ConfigMap, error) {
	bootstrap, err := cmd.Flags().GetString("bootstrap")
	if err != nil {
		return nil, err
	}
	caLocation, err := cmd.Flags().GetString("ca-location")
	if err != nil {
		return nil, err
	}
	configMap := getOnPremCommonConfig(clientID, bootstrap, caLocation)

	if err := configMap.SetKey("retry.backoff.ms", "250"); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("request.timeout.ms", "10000"); err != nil {
		return nil, err
	}

	if err := SetProducerDebugOption(configMap); err != nil {
		return nil, err
	}

	return setProtocolConfig(cmd, configMap)
}

func getOnPremConsumerConfigMap(cmd *cobra.Command, clientID string) (*ckafka.ConfigMap, error) {
	bootstrap, err := cmd.Flags().GetString("bootstrap")
	if err != nil {
		return nil, err
	}
	caLocation, err := cmd.Flags().GetString("ca-location")
	if err != nil {
		return nil, err
	}
	configMap := getOnPremCommonConfig(clientID, bootstrap, caLocation)

	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return nil, err
	}
	if group == "" {
		group = fmt.Sprintf("confluent_cli_consumer_%s", uuid.New())
	}
	if err := configMap.SetKey("group.id", group); err != nil {
		return nil, err
	}
	log.CliLogger.Debugf("Created consumer group: %s", group)

	// see explanation: https://www.confluent.io/blog/incremental-cooperative-rebalancing-in-kafka/
	if err := configMap.SetKey("partition.assignment.strategy", "cooperative-sticky"); err != nil {
		return nil, err
	}

	if err := SetConsumerDebugOption(configMap); err != nil {
		return nil, err
	}

	return setProtocolConfig(cmd, configMap)
}

func setProtocolConfig(cmd *cobra.Command, configMap *ckafka.ConfigMap) (*ckafka.ConfigMap, error) {
	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return nil, err
	}
	switch protocol {
	case "SSL":
		configMap, err = setSSLConfig(cmd, configMap)
		if err != nil {
			return nil, err
		}
	case "SASL_SSL":
		configMap, err = setSaslConfig(cmd, configMap)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.NewErrorWithSuggestions(fmt.Errorf(errors.InvalidSecurityProtocolErrorMsg, protocol).Error(), errors.OnPremConfigGuideSuggestions)
	}
	return configMap, nil
}

func setSSLConfig(cmd *cobra.Command, configMap *ckafka.ConfigMap) (*ckafka.ConfigMap, error) {
	certLocation, err := cmd.Flags().GetString("cert-location")
	if err != nil {
		return nil, err
	}
	keyLocation, err := cmd.Flags().GetString("key-location")
	if err != nil {
		return nil, err
	}
	keyPassword, err := cmd.Flags().GetString("key-password")
	if err != nil {
		return nil, err
	}
	sslMap := map[string]string{
		"security.protocol":        "SSL",
		"ssl.certificate.location": certLocation,
		"ssl.key.location":         keyLocation,
		"ssl.key.password":         keyPassword,
	}
	for key, value := range sslMap {
		if err := configMap.SetKey(key, value); err != nil {
			return nil, err
		}
	}
	return configMap, nil
}

func setSaslConfig(cmd *cobra.Command, configMap *ckafka.ConfigMap) (*ckafka.ConfigMap, error) {
	saslMechanism, err := cmd.Flags().GetString("sasl-mechanism")
	if err != nil {
		return nil, err
	}
	saslMap := map[string]string{
		"security.protocol": "SASL_SSL",
		"sasl.mechanism":    saslMechanism,
	}
	if saslMechanism == "PLAIN" {
		username, password, err := promptForSASLAuth(cmd)
		if err != nil {
			return nil, err
		}
		saslMap["sasl.username"] = username
		saslMap["sasl.password"] = password
	} else {
		saslMap["sasl.oauthbearer.config"] = oauthConfig
	}
	for key, value := range saslMap {
		if err := configMap.SetKey(key, value); err != nil {
			return nil, err
		}
	}
	return configMap, nil
}

func promptForSASLAuth(cmd *cobra.Command) (string, string, error) {
	username, err := cmd.Flags().GetString("username")
	if err != nil {
		return "", "", err
	}
	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return "", "", err
	}
	if username != "" && password != "" {
		return username, password, nil
	}
	f := form.New(
		form.Field{ID: "username", Prompt: "Enter your SASL username"},
		form.Field{ID: "password", Prompt: "Enter your SASL password", IsHidden: true},
	)
	if err := f.Prompt(form.NewPrompt()); err != nil {
		return "", "", err
	}
	return f.Responses["username"].(string), f.Responses["password"].(string), nil
}

func GetOffsetWithFallback(cmd *cobra.Command) (ckafka.Offset, error) {
	if cmd.Flags().Changed("offset") {
		offset, err := cmd.Flags().GetInt64("offset")
		if err != nil {
			return ckafka.OffsetInvalid, err
		}
		if offset < 0 {
			return ckafka.OffsetInvalid, errors.New(errors.InvalidOffsetErrorMsg)
		}
		return ckafka.NewOffset(offset)
	} else {
		fromBeginning, err := cmd.Flags().GetBool("from-beginning")
		if err != nil {
			return ckafka.OffsetInvalid, err
		}
		autoOffsetReset := "latest"
		if fromBeginning {
			autoOffsetReset = "earliest"
		}
		return ckafka.NewOffset(autoOffsetReset)
	}
}

func getPartitionsByIndex(partitions []ckafka.TopicPartition, partitionFilter PartitionFilter) []ckafka.TopicPartition {
	if partitionFilter.Changed {
		for _, partition := range partitions {
			if partition.Partition == partitionFilter.Index {
				log.CliLogger.Debugf("Consuming from partition: %d", partitionFilter.Index)
				return []ckafka.TopicPartition{partition}
			}
		}
		return []ckafka.TopicPartition{}
	}
	return partitions
}

func SetProducerDebugOption(configMap *ckafka.ConfigMap) error {
	switch log.CliLogger.Level {
	case log.DEBUG:
		return configMap.Set("debug=broker, topic, msg, protocol")
	case log.TRACE:
		return configMap.Set("debug=all")
	}
	return nil
}

func SetConsumerDebugOption(configMap *ckafka.ConfigMap) error {
	switch log.CliLogger.Level {
	case log.DEBUG:
		return configMap.Set("debug=broker, topic, msg, protocol, consumer, cgrp, fetch")
	case log.TRACE:
		return configMap.Set("debug=all")
	}
	return nil
}

func newProducerWithOverwrittenConfigs(configMap *ckafka.ConfigMap, configPath string, configStrings []string) (*ckafka.Producer, error) {
	if err := OverwriteKafkaClientConfigs(configMap, configPath, configStrings); err != nil {
		return nil, err
	}

	return ckafka.NewProducer(configMap)
}

func newConsumerWithOverwrittenConfigs(configMap *ckafka.ConfigMap, configPath string, configStrings []string) (*ckafka.Consumer, error) {
	if err := OverwriteKafkaClientConfigs(configMap, configPath, configStrings); err != nil {
		return nil, err
	}

	return ckafka.NewConsumer(configMap)
}

func OverwriteKafkaClientConfigs(configMap *ckafka.ConfigMap, configPath string, configs []string) error {
	configurations := make(map[string]string)
	if configPath != "" {
		configFile, err := os.Open(configPath)
		if err != nil {
			return err
		}
		defer configFile.Close()
		configBytes, err := io.ReadAll(configFile)
		if err != nil {
			return err
		}
		clientConfigs := &kafkaClientConfigs{}
		if err := json.Unmarshal(configBytes, &clientConfigs.configurations); err != nil {
			return err
		}
		configurations = clientConfigs.configurations
	}

	var err error
	if len(configs) > 0 {
		configurations, err = properties.ConfigFlagToMap(configs)
		if err != nil {
			return err
		}
	}

	for key, value := range configurations {
		if err := configMap.SetKey(key, value); err != nil {
			return err
		}
		log.CliLogger.Debugf(`Overwrote the value of client configuration "%s" to "%s"`, key, value)
	}

	return nil
}
