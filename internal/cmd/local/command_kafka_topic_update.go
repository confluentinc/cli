package local

import (
	"context"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *kafkaCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <topic>",
		Short: "Update a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.topicUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "my_topic" topic.`,
				Code: "confluent kafka topic describe my_topic",
			},
		),
	}

	cmd.Flags().StringSlice("config", nil, `A comma-separated list of topics configuration ("key=value") overrides for the topic being created.`)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *kafkaCommand) topicUpdate(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return err
	}

	topicName := args[0]

	return kafka.UpdateTopicWithRestClient(cmd, restClient, context.Background(), topicName, clusterId)
}
