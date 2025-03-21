package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/broker"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *command) newKafkaClusterConfigurationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local Kafka cluster configurations.",
		Args:  cobra.NoArgs,
		RunE:  c.configurationList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List configuration values for the Kafka cluster.",
				Code: "confluent local kafka cluster configuration list",
			},
		),
	}

	cmd.Flags().String("config", "", "Get a specific configuration value.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) configurationList(cmd *cobra.Command, _ []string) error {
	configName, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	clusterConfig, err := broker.GetClusterWideConfigs(restClient, context.Background(), clusterId, configName)
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
