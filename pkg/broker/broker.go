package broker

import (
	"context"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/errors"
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

func CheckAllOrIdSpecified(cmd *cobra.Command, args []string, checkAll bool) (int32, bool, error) {
	var all bool
	var err error
	if checkAll {
		if cmd.Flags().Changed("all") && len(args) > 0 {
			return -1, false, errors.New(errors.OnlySpecifyAllOrBrokerIDErrorMsg)
		}
		if !cmd.Flags().Changed("all") && len(args) == 0 {
			return -1, false, errors.New(errors.MustSpecifyAllOrBrokerIDErrorMsg)
		}
		all, err = cmd.Flags().GetBool("all")
		if err != nil {
			return -1, false, err
		}
	}
	if len(args) > 0 {
		brokerIdStr := args[0]
		brokerId, err := strconv.ParseInt(brokerIdStr, 10, 32)
		return int32(brokerId), false, err
	}
	return -1, all, nil
}

func ParseClusterConfigData(clusterConfig kafkarestv3.ClusterConfigDataList) []*ConfigOut {
	configs := make([]*ConfigOut, len(clusterConfig.Data))
	for i, data := range clusterConfig.Data {
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

func parseBrokerConfigData(brokerConfig kafkarestv3.BrokerConfigDataList) []*ConfigOut {
	configs := make([]*ConfigOut, len(brokerConfig.Data))
	for i, data := range brokerConfig.Data {
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

// getClusterWideConfigs fetches cluster-wide configs or just configName config if specified
func GetClusterWideConfigs(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId, configName string) (kafkarestv3.ClusterConfigDataList, error) {
	var clusterConfig kafkarestv3.ClusterConfigDataList
	var resp *http.Response
	var err error
	if configName != "" { // Get config specified by configName
		var configNameData kafkarestv3.ClusterConfigData
		configNameData, resp, err = restClient.ConfigsV3Api.GetKafkaClusterConfig(restContext, clusterId, configName)
		clusterConfig.Data = []kafkarestv3.ClusterConfigData{configNameData}
	} else { // Get all configs
		clusterConfig, resp, err = restClient.ConfigsV3Api.ListKafkaClusterConfigs(restContext, clusterId)
	}
	if err != nil {
		return clusterConfig, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}
	return clusterConfig, nil
}

// getIndividualBrokerConfigs fetches all per-broker configs or just the config specified by configName
func getIndividualBrokerConfigs(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, brokerId int32, configName string) (kafkarestv3.BrokerConfigDataList, error) {
	var brokerConfig kafkarestv3.BrokerConfigDataList
	var resp *http.Response
	var err error
	if configName != "" {
		var brokerNameData kafkarestv3.BrokerConfigData
		brokerNameData, resp, err = restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsNameGet(restContext, clusterId, brokerId, configName)
		brokerConfig.Data = []kafkarestv3.BrokerConfigData{brokerNameData}
	} else {
		brokerConfig, resp, err = restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsGet(restContext, clusterId, brokerId)
	}
	if err != nil {
		return brokerConfig, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}
	return brokerConfig, nil
}

func ToAlterConfigBatchRequestDataOnPrem(configsMap map[string]string) kafkarestv3.AlterConfigBatchRequestData {
	kafkaRestConfigs := make([]kafkarestv3.AlterConfigBatchRequestDataData, len(configsMap))
	i := 0
	for key, val := range configsMap {
		v := val
		kafkaRestConfigs[i] = kafkarestv3.AlterConfigBatchRequestDataData{
			Name:  key,
			Value: &v,
		}
		i++
	}
	return kafkarestv3.AlterConfigBatchRequestData{Data: kafkaRestConfigs}
}
