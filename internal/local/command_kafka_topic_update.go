package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/internal/kafka"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newKafkaTopicUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <topic>",
		Short: "Update a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.kafkaTopicUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "test" topic.`,
				Code: "confluent kafka topic describe test",
			},
		),
	}

	cmd.Flags().StringSlice("config", nil, `A comma-separated list of topics configuration ("key=value") overrides for the topic being created.`)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) kafkaTopicUpdate(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRestNotReadySuggestion)
	}

	return kafka.UpdateTopic(cmd, restClient, context.Background(), args[0], clusterId)
}
