package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/internal/kafka"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *Command) newKafkaTopicDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <topic-1> [topic-2] ... [topic-n]",
		Short: "Delete one or more Kafka topics.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.kafkaTopicDelete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete the topic "test". Use this command carefully as data loss can occur.`,
				Code: "confluent local kafka topic delete test",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *Command) kafkaTopicDelete(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	return kafka.DeleteTopic(cmd, restClient, context.Background(), args, clusterId)
}
