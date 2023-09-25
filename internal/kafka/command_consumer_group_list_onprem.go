package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newGroupListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumer groups.",
		Args:  cobra.NoArgs,
		RunE:  c.groupListOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) groupListOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	groups, resp, err := restClient.ConsumerGroupV3Api.ListKafkaConsumerGroups(restContext, clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, group := range groups.Data {
		list.Add(&consumerGroupOut{
			ClusterId:         group.ClusterId,
			ConsumerGroupId:   group.ConsumerGroupId,
			Coordinator:       getStringBroker(group.Coordinator.Related),
			IsSimple:          group.IsSimple,
			PartitionAssignor: group.PartitionAssignor,
			State:             group.State,
		})
	}
	return list.Print()
}
