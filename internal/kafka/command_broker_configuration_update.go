package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *brokerCommand) newConfigurationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update Kafka broker configurations.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.configurationUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update configuration values for broker 1.",
				Code: "confluent kafka broker configuration update 1 --config min.insync.replicas=2,num.partitions=2",
			},
		),
	}

	pcmd.AddConfigFlag(cmd)
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("config"))

	return cmd
}

func (c *brokerCommand) configurationUpdate(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	brokerId, err := broker.GetId(cmd, args)
	if err != nil {
		return err
	}

	configs, err := broker.Update(cmd, args, restClient, restContext, clusterId)
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
