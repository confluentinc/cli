package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/broker"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newKafkaBrokerConfigurationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <id>",
		Short: "List local Kafka broker configurations.",
		Long:  "Describe per-broker configuration values.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.brokerConfigurationList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "min.insync.replicas" configuration for broker 1.`,
				Code: "confluent local broker configuration list 1 --config min.insync.replicas",
			},
		),
	}

	cmd.Flags().String("config", "", "Get a specific configuration value.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) brokerConfigurationList(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	return broker.ConfigurationList(cmd, args, restClient, context.Background(), clusterId)
}
