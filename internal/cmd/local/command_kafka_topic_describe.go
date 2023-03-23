package local

import (
	"context"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *kafkaCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <topic>",
		Args:  cobra.ExactArgs(1),
		Short: "Describe a Kafka topic.",
		RunE:  c.topicDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "my_topic" topic.`,
				Code: "confluent local kafka topic describe my_topic",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *kafkaCommand) topicDescribe(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return err
	}

	return kafka.DescribeTopicWithRESTClient(cmd, restClient, context.Background(), args[0], clusterId)
}
