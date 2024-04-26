package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func (c *clusterCommand) newConfigurationListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cluster-wide Kafka broker configurations.",
		Args:  cobra.NoArgs,
		RunE:  c.configurationUpdateOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List configuration values for all brokers in the cluster.",
				Code: "confluent kafka broker list",
			},
		),
	}

	cmd.Flags().String("config", "", "Get a specific configuration value.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *clusterCommand) configurationListOnPrem(cmd *cobra.Command, _ []string) error {
	configName, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterConfig, err := broker.GetClusterWideConfigs(restClient, restContext, clusterId, configName)
	if err != nil {
		return err
	}
	configs := broker.ParseClusterConfigData(clusterConfig)

	list := output.NewList(cmd)
	for _, config := range configs {
		if output.GetFormat(cmd) == output.Human {
			config.Name = utils.Abbreviate(config.Name, broker.AbbreviationLength)
			config.Value = utils.Abbreviate(config.Value, broker.AbbreviationLength)
		}
		list.Add(config)
	}
	return list.Print()
}
