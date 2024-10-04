package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/broker"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newKafkaBrokerConfigurationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update local Kafka broker configurations.",
		Long:  "Update per-broker configurations for local Kafka brokers.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.brokerConfigurationUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update configuration values for broker 1.",
				Code: "confluent kafka broker configuration update 1 --config min.insync.replicas=2,num.partitions=2",
			},
		),
	}

	pcmd.AddConfigFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("config"))

	return cmd
}

func (c *command) brokerConfigurationUpdate(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	brokerId, err := broker.GetId(cmd, args)
	if err != nil {
		return err
	}

	configs, err := broker.Update(cmd, args, restClient, context.Background(), clusterId)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		output.Printf(c.Config.EnableColor, "Updated the following configurations for broker \"%d\":\n", brokerId)
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
