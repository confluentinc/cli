package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newKafkaTopicCreateCommand() *cobra.Command {
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
	cmd.Flags().StringSlice("config", nil, `A comma-separated list of topic configuration ("key=value") overrides for the topic being created.`)
	cmd.Flags().Bool("if-not-exists", false, "Exit gracefully if topic already exists.")

	return cmd
}

func (c *command) kafkaTopicCreate(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return err
	}

	return kafka.CreateTopicWithRestClient(cmd, restClient, context.Background(), args[0], clusterId)
}
