package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *Command) newKafkaTopicCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <topic>",
		Short: "Create a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kafkaTopicCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a topic named "test" with specified configuration parameters.`,
				Code: "confluent local kafka topic create test --config cleanup.policy=compact,compression.type=gzip",
			},
		),
	}

	cmd.Flags().Uint32("partitions", 0, "Number of topic partitions.")
	cmd.Flags().Uint32("replication-factor", 0, "Number of replicas.")
	pcmd.AddConfigFlag(cmd)
	cmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")

	return cmd
}

func (c *Command) kafkaTopicCreate(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	return kafka.CreateTopic(cmd, restClient, context.Background(), args[0], clusterId)
}
