package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (c *command) newKafkaTopicListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List local Kafka topics.",
		Args:  cobra.NoArgs,
		RunE:  c.kafkaTopicList,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) kafkaTopicList(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), kafkaRESTNotReadySuggestion)
	}

	return kafka.ListTopicsWithRestClient(cmd, restClient, context.Background(), clusterId)
}
