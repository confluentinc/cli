package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/internal/kafka"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newKafkaTopicConfigurationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <topic>",
		Args:  cobra.ExactArgs(1),
		Short: "List Kafka topic configurations.",
		RunE:  c.kafkaTopicConfigurationList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List configurations for topic "test".`,
				Code: "confluent local kafka topic configuration list test",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) kafkaTopicConfigurationList(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	return kafka.ListConfigurations(cmd, restClient, context.Background(), args[0], clusterId)
}
