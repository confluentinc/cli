package broker

import (
	"context"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
)

func Update(cmd *cobra.Command, args []string, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string, checkAll bool) error {
	brokerId, all, err := CheckAllOrIdSpecified(cmd, args, checkAll)
	if err != nil {
		return err
	}

	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configMap, err := properties.GetMap(config)
	if err != nil {
		return err
	}
	configs := ToAlterConfigBatchRequestDataOnPrem(configMap)

	if all {
		opts := &kafkarestv3.UpdateKafkaClusterConfigsOpts{AlterConfigBatchRequestData: optional.NewInterface(configs)}
		if resp, err := restClient.ConfigsV3Api.UpdateKafkaClusterConfigs(restContext, clusterId, opts); err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
	} else {
		opts := &kafkarestv3.ClustersClusterIdBrokersBrokerIdConfigsalterPostOpts{AlterConfigBatchRequestData: optional.NewInterface(configs)}
		if resp, err := restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsalterPost(restContext, clusterId, brokerId, opts); err != nil {
			return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
		}
	}

	if output.GetFormat(cmd) == output.Human {
		if all {
			output.Printf(false, "Updated the following broker configurations for cluster \"%s\":\n", clusterId)
		} else {
			output.Printf(false, "Updated the following configurations for broker \"%d\":\n", brokerId)
		}
	}

	list := output.NewList(cmd)
	for _, config := range configs.Data {
		list.Add(&ConfigOut{
			Name:  config.Name,
			Value: *config.Value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
