package kafka

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	ckgo "github.com/confluentinc/confluent-kafka-go/v2/kafka"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/form"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/properties"
)

type kafkaClientConfigs struct {
	configurations map[string]string
}

type PartitionFilter struct {
	Changed bool
	Index   int32
}

func getCommonConfig(kafka *config.KafkaClusterConfig, clientId string) (*ckgo.ConfigMap, error) {
	if err := kafka.DecryptAPIKeys(); err != nil {
		return nil, err
	}

	configMap := &ckgo.ConfigMap{
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

func getProducerConfigMap(kafka *config.KafkaClusterConfig, clientID string) (*ckgo.ConfigMap, error) {
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

func getConsumerConfigMap(group string, kafka *config.KafkaClusterConfig, clientID string) (*ckgo.ConfigMap, error) {
	configMap, err := getCommonConfig(kafka, clientID)
	if err != nil {
		return nil, err
	}
	if err := configMap.SetKey("group.id", group); err != nil {
		return nil, err
	}
	log.CliLogger.Debugf("Created consumer group: %s", group)

	if err := configMap.SetKey("enable.auto.commit", false); err != nil {
		return nil, err
	}

	// see explanation: https://www.confluent.io/blog/incremental-cooperative-rebalancing-in-kafka/
	if err := configMap.SetKey("partition.assignment.strategy", "cooperative-sticky"); err != nil {
		return nil, err
	}
	if err := SetConsumerDebugOption(configMap); err != nil {
		return nil, err
	}
	return configMap, nil
}

func getOnPremCommonConfig(clientID, bootstrap string) *ckgo.ConfigMap {
	return &ckgo.ConfigMap{
		"ssl.endpoint.identification.algorithm": "https",
		"client.id":                             clientID,
		"bootstrap.servers":                     bootstrap,
	}
}

func getOnPremProducerConfigMap(cmd *cobra.Command, clientID string) (*ckgo.ConfigMap, error) {
	bootstrap, err := cmd.Flags().GetString("bootstrap")
	if err != nil {
		return nil, err
	}

	configMap := getOnPremCommonConfig(clientID, bootstrap)

	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return nil, err
	}
	if protocol == "SSL" || protocol == "SASL_SSL" {
		certificateAuthorityPath, err := cmd.Flags().GetString("certificate-authority-path")
		if err != nil {
			return nil, err
		}

		if err := configMap.SetKey("enable.ssl.certificate.verification", true); err != nil {
			return nil, err
		}
		if err := configMap.SetKey("ssl.ca.location", certificateAuthorityPath); err != nil {
			return nil, err
		}
	}

	if err := configMap.SetKey("retry.backoff.ms", 250); err != nil {
		return nil, err
	}
	if err := configMap.SetKey("request.timeout.ms", 10000); err != nil {
		return nil, err
	}

	if err := SetProducerDebugOption(configMap); err != nil {
		return nil, err
	}

	return setProtocolConfig(cmd, configMap)
}

func getOnPremConsumerConfigMap(cmd *cobra.Command, clientID string) (*ckgo.ConfigMap, error) {
	bootstrap, err := cmd.Flags().GetString("bootstrap")
	if err != nil {
		return nil, err
	}

	configMap := getOnPremCommonConfig(clientID, bootstrap)

	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return nil, err
	}
	if protocol == "SSL" || protocol == "SASL_SSL" {
		certificateAuthorityPath, err := cmd.Flags().GetString("certificate-authority-path")
		if err != nil {
			return nil, err
		}

		if err := configMap.SetKey("enable.ssl.certificate.verification", true); err != nil {
			return nil, err
		}
		if err := configMap.SetKey("ssl.ca.location", certificateAuthorityPath); err != nil {
			return nil, err
		}
	}

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

func setProtocolConfig(cmd *cobra.Command, configMap *ckgo.ConfigMap) (*ckgo.ConfigMap, error) {
	protocol, err := cmd.Flags().GetString("protocol")
	if err != nil {
		return nil, err
	}
	switch protocol {
	case "PLAINTEXT":
		configMap, err = setPlaintextConfig(configMap)
		if err != nil {
			return nil, err
		}
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
		return nil, errors.NewErrorWithSuggestions(
			fmt.Errorf("security protocol not supported: %s", protocol).Error(),
			errors.OnPremConfigGuideSuggestions,
		)
	}
	return configMap, nil
}

func setPlaintextConfig(configMap *ckgo.ConfigMap) (*ckgo.ConfigMap, error) {
	if err := configMap.SetKey("security.protocol", "PLAINTEXT"); err != nil {
		return nil, err
	}
	return configMap, nil
}

func setSSLConfig(cmd *cobra.Command, configMap *ckgo.ConfigMap) (*ckgo.ConfigMap, error) {
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

func setSaslConfig(cmd *cobra.Command, configMap *ckgo.ConfigMap) (*ckgo.ConfigMap, error) {
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

func GetOffsetWithFallback(cmd *cobra.Command) (ckgo.Offset, error) {
	if cmd.Flags().Changed("offset") {
		offset, err := cmd.Flags().GetInt64("offset")
		if err != nil {
			return ckgo.OffsetInvalid, err
		}
		if offset < 0 {
			return ckgo.OffsetInvalid, fmt.Errorf("offset value must be a non-negative integer")
		}
		return ckgo.NewOffset(offset)
	} else {
		fromBeginning, err := cmd.Flags().GetBool("from-beginning")
		if err != nil {
			return ckgo.OffsetInvalid, err
		}
		autoOffsetReset := "latest"
		if fromBeginning {
			autoOffsetReset = "earliest"
		}
		return ckgo.NewOffset(autoOffsetReset)
	}
}

func getPartitionsByIndex(partitions []ckgo.TopicPartition, partitionFilter PartitionFilter) []ckgo.TopicPartition {
	if partitionFilter.Changed {
		for _, partition := range partitions {
			if partition.Partition == partitionFilter.Index {
				log.CliLogger.Debugf("Consuming from partition: %d", partitionFilter.Index)
				return []ckgo.TopicPartition{partition}
			}
		}
		return []ckgo.TopicPartition{}
	}
	return partitions
}

func SetProducerDebugOption(configMap *ckgo.ConfigMap) error {
	switch log.CliLogger.Level {
	case log.DEBUG:
		return configMap.Set("debug=broker, topic, msg, protocol")
	case log.TRACE, log.UNSAFE_TRACE:
		return configMap.Set("debug=all")
	}
	return nil
}

func SetConsumerDebugOption(configMap *ckgo.ConfigMap) error {
	switch log.CliLogger.Level {
	case log.DEBUG:
		return configMap.Set("debug=broker, topic, msg, protocol, consumer, cgrp, fetch")
	case log.TRACE, log.UNSAFE_TRACE:
		return configMap.Set("debug=all")
	}
	return nil
}

func newProducerWithOverwrittenConfigs(configMap *ckgo.ConfigMap, configPath string, configStrings []string) (*ckgo.Producer, error) {
	if err := OverwriteKafkaClientConfigs(configMap, configPath, configStrings); err != nil {
		return nil, err
	}
	log.CliLogger.Debug("Creating Confluent Kafka producer with the configuration map.")
	return ckgo.NewProducer(configMap)
}

func newConsumerWithOverwrittenConfigs(configMap *ckgo.ConfigMap, configPath string, configStrings []string) (*ckgo.Consumer, error) {
	if err := OverwriteKafkaClientConfigs(configMap, configPath, configStrings); err != nil {
		return nil, err
	}
	log.CliLogger.Debug("Creating Confluent Kafka consumer with the configuration map.")
	return ckgo.NewConsumer(configMap)
}

func OverwriteKafkaClientConfigs(configMap *ckgo.ConfigMap, configPath string, configs []string) error {
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
