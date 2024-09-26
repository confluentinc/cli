package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/broker"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *brokerCommand) newConfigurationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <id>",
		Short: "List Kafka broker configurations.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.configurationList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all configurations for broker 1.",
				Code: "confluent kafka broker configuration list 1",
			},
			examples.Example{
				Text: `Describe the "min.insync.replicas" configuration for broker 1.`,
				Code: "confluent kafka broker configuration list 1 --config min.insync.replicas",
			},
		),
	}

	cmd.Flags().String("config", "", "Get a specific configuration value.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *brokerCommand) configurationList(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return broker.ConfigurationList(cmd, args, restClient, restContext, clusterId)
}
