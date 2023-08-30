package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newKafkaBrokerDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe [id]",
		Short: "Describe a local Kafka broker.",
		Long:  "Describe per-broker configuration values.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.brokerDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "min.insync.replicas" configuration for broker 1.`,
				Code: "confluent local broker describe 1 --config-name min.insync.replicas",
			},
		),
	}

	cmd.Flags().String("config-name", "", "Get a specific configuration value.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) brokerDescribe(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	return broker.Describe(cmd, args, restClient, context.Background(), clusterId, false)
}
