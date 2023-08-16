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

func (c *Command) newKafkaBrokerDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe [id]",
		Short: "Describe a local Kafka broker.",
		Long:  "Describe cluster-wide or per-broker configuration values.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.kafkaDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "min.insync.replicas" configuration for broker 1.`,
				Code: "confluent local broker describe 1 --config-name min.insync.replicas",
			},
			examples.Example{
				Text: "Describe the non-default cluster-wide broker configuration values.",
				Code: "confluent local broker describe --all",
			},
		),
	}

	cmd.Flags().Bool("all", false, "Get cluster-wide broker configurations (non-default values only).")
	cmd.Flags().String("config-name", "", "Get a specific configuration value (pair with --all to see a cluster-wide configuration.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *Command) kafkaDescribe(cmd *cobra.Command, args []string) error {
	brokerId, all, err := broker.CheckAllOrIdSpecified(cmd, args)
	if err != nil {
		return err
	}

	configName, err := cmd.Flags().GetString("config-name")
	if err != nil {
		return err
	}

	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	// Get Broker Configs
	var data []*broker.ConfigOut
	if all { // fetch cluster-wide configs
		clusterConfig, err := broker.GetClusterWideConfigs(restClient, context.Background(), clusterId, configName)
		if err != nil {
			return err
		}
		data = broker.ParseClusterConfigData(clusterConfig)
	} else { // fetch individual broker configs
		brokerConfig, err := broker.GetIndividualBrokerConfigs(restClient, context.Background(), clusterId, brokerId, configName)
		if err != nil {
			return err
		}
		data = broker.ParseBrokerConfigData(brokerConfig)
	}

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
