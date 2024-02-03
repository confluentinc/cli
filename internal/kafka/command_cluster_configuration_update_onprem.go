package kafka

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/properties"
)

func (c *clusterCommand) newConfigurationUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update cluster-wide Kafka broker configurations.",
		Args:  cobra.NoArgs,
		RunE:  c.configurationUpdateOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update configuration values for all brokers in the cluster.",
				Code: "confluent kafka broker update --config min.insync.replicas=2,num.partitions=2",
			},
		),
	}

	pcmd.AddConfigFlag(cmd)
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("config"))

	return cmd
}

func (c *clusterCommand) configurationUpdateOnPrem(cmd *cobra.Command, _ []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
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
	configs := broker.ToAlterConfigBatchRequestDataOnPrem(configMap)

	opts := &kafkarestv3.UpdateKafkaClusterConfigsOpts{AlterConfigBatchRequestData: optional.NewInterface(configs)}
	if resp, err := restClient.ConfigsV3Api.UpdateKafkaClusterConfigs(restContext, clusterId, opts); err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	if output.GetFormat(cmd) == output.Human {
		output.Printf(c.Config.EnableColor, "Updated the following broker configurations for cluster \"%s\":\n", clusterId)
	}

	list := output.NewList(cmd)
	for _, config := range configs.Data {
		list.Add(&broker.ConfigOut{
			Name:  config.Name,
			Value: *config.Value,
		})
	}
	list.Filter([]string{"Name", "Value"})
	return list.Print()
}
