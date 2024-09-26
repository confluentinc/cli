package broker

import (
	"context"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/properties"
)

func Update(cmd *cobra.Command, args []string, restClient *kafkarestv3.APIClient, restContext context.Context, clusterId string) (kafkarestv3.AlterConfigBatchRequestData, error) {
	brokerId, err := GetId(cmd, args)
	if err != nil {
		return kafkarestv3.AlterConfigBatchRequestData{}, err
	}

	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return kafkarestv3.AlterConfigBatchRequestData{}, err
	}
	configMap, err := properties.GetMap(config)
	if err != nil {
		return kafkarestv3.AlterConfigBatchRequestData{}, err
	}
	configs := ToAlterConfigBatchRequestDataOnPrem(configMap)

	opts := &kafkarestv3.ClustersClusterIdBrokersBrokerIdConfigsalterPostOpts{AlterConfigBatchRequestData: optional.NewInterface(configs)}
	if resp, err := restClient.ConfigsV3Api.ClustersClusterIdBrokersBrokerIdConfigsalterPost(restContext, clusterId, brokerId, opts); err != nil {
		return kafkarestv3.AlterConfigBatchRequestData{}, kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	return configs, nil
}
