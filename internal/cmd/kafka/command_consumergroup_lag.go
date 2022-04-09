package kafka

import (
	cloudkafkarest "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

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

	c.AddCommand(c.newGetCommand())
	c.AddCommand(c.newListCommand())
	c.AddCommand(c.newSummarizeCommand())

	return c.Command
}

func convertLagToStruct(lagData cloudkafkarest.ConsumerLagData) *lagDataStruct {
	instanceId := ""
	if lagData.InstanceId.IsSet() {
		instanceId = *lagData.InstanceId.Get()
	}

	return &lagDataStruct{
		ClusterId:       lagData.ClusterId,
		ConsumerGroupId: lagData.ConsumerGroupId,
		Lag:             lagData.Lag,
		LogEndOffset:    lagData.LogEndOffset,
		CurrentOffset:   lagData.CurrentOffset,
		ConsumerId:      lagData.ConsumerId,
		InstanceId:      instanceId,
		ClientId:        lagData.ClientId,
		TopicName:       lagData.TopicName,
		PartitionId:     lagData.PartitionId,
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
