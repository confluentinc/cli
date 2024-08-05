package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
)

func (c *command) newKafkaBrokerDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a local Kafka broker.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.brokerDescribe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) brokerDescribe(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	return broker.Describe(cmd, args, restClient, context.Background(), clusterId)
}
