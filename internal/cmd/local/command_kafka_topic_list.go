package local

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *kafkaCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list Local kafka topics",
		Args:  cobra.NoArgs,
		RunE:  c.topicList,
	}

	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *kafkaCommand) topicList(cmd *cobra.Command, args []string) error {
	restClient, clusterId, err := initKafkaRest(c.CLICommand, cmd)
	if err != nil {
		return err
	}

	return kafka.ListTopicsWithRESTClient(cmd, restClient, context.Background(), clusterId)
}
