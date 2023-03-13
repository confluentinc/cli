package kafka

import (
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type lagDataStruct struct {
	ClusterId       string `human:"Cluster" serialized:"cluster"`
	ConsumerGroupId string `human:"Consumer Group" serialized:"consumer_group"`
	Lag             int64  `human:"Lag" serialized:"lag"`
	LogEndOffset    int64  `human:"Log End Offset" serialized:"log_end_offset"`
	CurrentOffset   int64  `human:"Current Offset" serialized:"current_offset"`
	ConsumerId      string `human:"Consumer" serialized:"consumer"`
	InstanceId      string `human:"Instance" serialized:"instance"`
	ClientId        string `human:"Client" serialized:"client"`
	TopicName       string `human:"Topic" serialized:"topic"`
	PartitionId     int32  `human:"Partition" serialized:"partition"`
}

type lagCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newLagCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "lag",
		Short:  "View consumer lag.",
		Hidden: true,
	}

	c := &lagCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	cmd.AddCommand(c.newGetCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newSummarizeCommand())

	return cmd
}

func convertLagToStruct(data kafkarestv3.ConsumerLagData) *lagDataStruct {
	return &lagDataStruct{
		ClusterId:       data.GetClusterId(),
		ConsumerGroupId: data.GetConsumerGroupId(),
		Lag:             data.GetLag(),
		LogEndOffset:    data.GetLogEndOffset(),
		CurrentOffset:   data.GetCurrentOffset(),
		ConsumerId:      data.GetConsumerId(),
		InstanceId:      data.GetInstanceId(),
		ClientId:        data.GetClientId(),
		TopicName:       data.GetTopicName(),
		PartitionId:     data.GetPartitionId(),
	}
}

func (c *lagCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConsumerGroups()
}

func (c *lagCommand) autocompleteConsumerGroups() []string {
	consumerGroupDataList, err := listConsumerGroups(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(consumerGroupDataList.Data))
	for i, consumerGroup := range consumerGroupDataList.Data {
		suggestions[i] = consumerGroup.ConsumerGroupId
	}
	return suggestions
}
