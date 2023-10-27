package local

import (
	"context"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
)

func (c *command) newKafkaClusterConfigurationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update local Kafka cluster configurations.",
		Args:  cobra.NoArgs,
		RunE:  c.configurationUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update configuration values for the Kafka cluster.",
				Code: "confluent local kafka cluster configuration update --config min.insync.replicas=2,num.partitions=2",
			},
		),
	}

	pcmd.AddConfigFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) configurationUpdate(cmd *cobra.Command, _ []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	config, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}
	configMap, err := properties.GetMap(config)
	if err != nil {
		return err
	}
	configs := broker.ToAlterConfigBatchRequestDataOnPrem(configMap)

	opts := &kafkarestv3.UpdateKafkaClusterConfigsOpts{AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{Data: configs})}
	if resp, err := restClient.ConfigsV3Api.UpdateKafkaClusterConfigs(context.Background(), clusterId, opts); err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	if output.GetFormat(cmd) == output.Human {
		output.Printf(c.Config.EnableColor, "Updated the following broker configurations for the Kafka cluster:\n")
	}

	list := output.NewList(cmd)
	for _, config := range configs {
		list.Add(&broker.ConfigOut{
			Name:  config.Name,
			Value: *config.Value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
