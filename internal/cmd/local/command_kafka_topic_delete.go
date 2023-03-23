package local

import (
	"context"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *localCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <topic>",
		Short: "Delete a Kafka topic.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.topicDelete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Delete the topic "my_topic". Use this command carefully as data loss can occur.`,
				Code: "confluent local kafka topic delete my_topic",
			},
		),
	}

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *localCommand) topicDelete(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return err
	}

	topicName := args[0]

	return kafka.DeleteTopicWithRestClient(cmd, restClient, context.Background(), topicName, clusterId)
}
