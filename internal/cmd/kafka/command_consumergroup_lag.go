package kafka

import (
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type lagCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	completableChildren []*cobra.Command
	*consumerGroupCommand
}

func NewLagCommand(prerunner pcmd.PreRunner, groupCmd *consumerGroupCommand) *lagCommand {
	cmd := &cobra.Command{
		Use:    "lag",
		Short:  "View consumer lag.",
		Hidden: true,
	}

	c := &lagCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner),
		consumerGroupCommand:          groupCmd,
	}

	summarizeCmd := c.newSummarizeCommand()
	listCmd := c.newListCommand()
	getCmd := c.newGetCommand()

	c.AddCommand(summarizeCmd)
	c.AddCommand(listCmd)
	c.AddCommand(getCmd)

	c.completableChildren = []*cobra.Command{summarizeCmd, listCmd, getCmd}

	return c
}

func convertLagToStruct(lagData kafkarestv3.ConsumerLagData) *lagDataStruct {
	instanceId := ""
	if lagData.InstanceId != nil {
		instanceId = *lagData.InstanceId
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
