package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/broker"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
)

func (c *Command) newKafkaBrokerListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local Kafka brokers.",
		Args:  cobra.NoArgs,
		RunE:  c.brokerList,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *Command) brokerList(cmd *cobra.Command, _ []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	return broker.List(cmd, restClient, context.Background(), clusterId)
}
