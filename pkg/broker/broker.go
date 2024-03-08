package broker

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/kafkarest"
)

const AbbreviationLength = 25

type ConfigOut struct {
	Name        string `human:"Name" serialized:"name"`
	Value       string `human:"Value,omitempty" serialized:"value,omitempty"`
	IsDefault   bool   `human:"Default" serialized:"is_default"`
	IsReadOnly  bool   `human:"Read-Only" serialized:"is_read_only"`
	IsSensitive bool   `human:"Sensitive" serialized:"is_sensitive"`
}

type out struct {
	ClusterId string `human:"Cluster" serialized:"cluster_id"`
	BrokerId  int32  `human:"Broker ID" serialized:"broker_id"`
	Host      string `human:"Host" serialized:"host"`
	Port      int32  `human:"Port" serialized:"port"`
}

func GetId(cmd *cobra.Command, args []string) (int32, error) {
	if len(args) > 0 {
		brokerId, err := strconv.ParseInt(args[0], 10, 32)
		return int32(brokerId), err
	}
	return -1, nil
}

func ParseClusterConfigData(clusterConfig []kafkarestv3.ClusterConfigData) []*ConfigOut {
	configs := make([]*ConfigOut, len(clusterConfig))
	for i, data := range clusterConfig {
		configs[i] = &ConfigOut{
			Name:        data.Name,
			IsDefault:   data.IsDefault,
			IsReadOnly:  data.IsReadOnly,
			IsSensitive: data.IsSensitive,
		}
		if data.Value != nil {
			configs[i].Value = *data.Value
		}
	}
	return configs
}

func parseBrokerConfigData(configs []kafkarestv3.BrokerConfigData) []*ConfigOut {
	out := make([]*ConfigOut, len(configs))
	for i, config := range configs {
		out[i] = &ConfigOut{
			Name:        config.Name,
			IsDefault:   config.IsDefault,
			IsReadOnly:  config.IsReadOnly,
			IsSensitive: config.IsSensitive,
		}
		if config.Value != nil {
			out[i].Value = *config.Value
		}
	}
	return out
}

// GetClusterWideConfigs fetches cluster-wide configs or just configName config if specified
func GetClusterWideConfigs(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId, configName string) ([]kafkarestv3.ClusterConfigData, error) {
	if configName != "" {
		config, resp, err := restClient.ConfigsV3Api.GetKafkaClusterConfig(restContext, clusterId, configName)
		if err != nil {
			return nil, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return []kafkarestv3.ClusterConfigData{config}, nil
	} else {
		configs, resp, err := restClient.ConfigsV3Api.ListKafkaClusterConfigs(restContext, clusterId)
		if err != nil {
			return nil, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return configs.Data, nil
	}
}

// getIndividualBrokerConfigs fetches all per-broker configs or just the config specified by configName
func getIndividualBrokerConfigs(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, brokerId int32, configName string) ([]kafkarestv3.BrokerConfigData, error) {
	if configName != "" {
		brokerNameData, resp, err := restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsNameGet(restContext, clusterId, brokerId, configName)
		if err != nil {
			return nil, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return []kafkarestv3.BrokerConfigData{brokerNameData}, nil
	} else {
		brokerConfig, resp, err := restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsGet(restContext, clusterId, brokerId)
		if err != nil {
			return nil, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
		return brokerConfig.Data, nil
	}
}

func ToAlterConfigBatchRequestDataOnPrem(configsMap map[string]string) kafkarestv3.AlterConfigBatchRequestData {
	configs := make([]kafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for key, val := range configsMap {
		val := val
		configs[i] = kafkarestv3.AlterConfigBatchRequestDataData{
			Name:  key,
			Value: &val,
		}
		i++
	}
	return kafkarestv3.AlterConfigBatchRequestData{Data: configs}
}
