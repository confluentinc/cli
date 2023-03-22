package local

import (
	"context"

	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"

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

	topicList, resp, err := restClient.TopicV3Api.ListKafkaTopics(context.Background(), clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, topic := range topicList.Data {
		list.Add(&topicOut{Name: topic.TopicName})
	}
	return list.Print()
}
