package kafka

import (
	"context"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const abbreviationLength = 25

type configOut struct {
	Name        string `human:"Name" serialized:"name"`
	Value       string `human:"Value,omitempty" serialized:"value,omitempty"`
	IsDefault   bool   `human:"Default" serialized:"is_default"`
	IsReadOnly  bool   `human:"Read Only" serialized:"is_read_only"`
	IsSensitive bool   `human:"Sensitive" serialized:"is_sensitive"`
}

func (c *brokerCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe [id]",
		Short: "Describe a Kafka broker.",
		Long:  "Describe cluster-wide or per-broker configuration values using Confluent Kafka REST.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "min.insync.replicas" configuration for broker 1.`,
				Code: "confluent kafka broker describe 1 --config-name min.insync.replicas",
			},
			examples.Example{
				Text: "Describe the non-default cluster-wide broker configuration values.",
				Code: "confluent kafka broker describe --all",
			},
		),
	}

	cmd.Flags().Bool("all", false, "Get cluster-wide broker configurations (non-default values only).")
	cmd.Flags().String("config-name", "", `Get a specific configuration value (pair with "--all" to see a cluster-wide config).`)
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *brokerCommand) describe(cmd *cobra.Command, args []string) error {
	brokerId, all, err := checkAllOrBrokerIdSpecified(cmd, args)
	if err != nil {
		return err
	}

	configName, err := cmd.Flags().GetString("config-name")
	if err != nil {
		return err
	}

	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	// Get Broker Configs
	var data []*configOut
	if all { // fetch cluster-wide configs
		clusterConfig, err := getClusterWideConfigs(restClient, restContext, clusterId, configName)
		if err != nil {
			return err
		}
		data = parseClusterConfigData(clusterConfig)
	} else { // fetch individual broker configs
		brokerConfig, err := getIndividualBrokerConfigs(restClient, restContext, clusterId, brokerId, configName)
		if err != nil {
			return err
		}
		data = parseBrokerConfigData(brokerConfig)
	}

	list := output.NewList(cmd)
	for _, entry := range data {
		if output.GetFormat(cmd) == output.Human {
			entry.Name = utils.Abbreviate(entry.Name, abbreviationLength)
			entry.Value = utils.Abbreviate(entry.Value, abbreviationLength)
		}
		list.Add(entry)
	}
	return list.Print()
}

func parseBrokerConfigData(brokerConfig kafkarestv3.BrokerConfigDataList) []*configOut {
	configs := make([]*configOut, len(brokerConfig.Data))
	for i, data := range brokerConfig.Data {
		configs[i] = &configOut{
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

func parseClusterConfigData(clusterConfig kafkarestv3.ClusterConfigDataList) []*configOut {
	configs := make([]*configOut, len(clusterConfig.Data))
	for i, data := range clusterConfig.Data {
		configs[i] = &configOut{
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

// getClusterWideConfigs fetches cluster-wide configs or just configName config if specified
func getClusterWideConfigs(restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, configName string) (kafkarestv3.ClusterConfigDataList, error) {
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
