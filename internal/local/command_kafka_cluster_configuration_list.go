package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func (c *command) newKafkaClusterConfigurationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local Kafka cluster configurations.",
		Args:  cobra.NoArgs,
		RunE:  c.configurationDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List configuration values for the Kafka cluster.",
				Code: "confluent local kafka cluster configuration list",
			},
		),
	}

	cmd.Flags().String("config-name", "", "Get a specific configuration value.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) configurationDescribe(cmd *cobra.Command, args []string) error {
	configName, err := cmd.Flags().GetString("config-name")
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
	data := broker.ParseClusterConfigData(clusterConfig)

	list := output.NewList(cmd)
	for _, entry := range data {
		if output.GetFormat(cmd) == output.Human {
			entry.Name = utils.Abbreviate(entry.Name, broker.AbbreviationLength)
			entry.Value = utils.Abbreviate(entry.Value, broker.AbbreviationLength)
		}
		list.Add(entry)
	}
	return list.Print()
}
