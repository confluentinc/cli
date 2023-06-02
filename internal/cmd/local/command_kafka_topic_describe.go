package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newKafkaTopicDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <topic>",
		Args:  cobra.ExactArgs(1),
		Short: "Describe a Kafka topic.",
		RunE:  c.kafkaTopicDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "test" topic.`,
				Code: "confluent local kafka topic describe test",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) kafkaTopicDescribe(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	return kafka.DescribeTopic(cmd, restClient, context.Background(), args[0], clusterId)
}
