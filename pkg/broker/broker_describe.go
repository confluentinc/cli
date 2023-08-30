package broker

import (
	"context"

	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

func Describe(cmd *cobra.Command, args []string, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, checkAll bool) error {
	brokerId, all, err := CheckAllOrIdSpecified(cmd, args, checkAll)
	if err != nil {
		return err
	}

	configName, err := cmd.Flags().GetString("config-name")
	if err != nil {
		return err
	}

	// Get Broker Configs
	var data []*ConfigOut
	if all { // fetch cluster-wide configs
		clusterConfig, err := GetClusterWideConfigs(restClient, restContext, clusterId, configName)
		if err != nil {
			return err
		}
		data = ParseClusterConfigData(clusterConfig)
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
			entry.Name = utils.Abbreviate(entry.Name, AbbreviationLength)
			entry.Value = utils.Abbreviate(entry.Value, AbbreviationLength)
		}
		list.Add(entry)
	}
	return list.Print()
}
