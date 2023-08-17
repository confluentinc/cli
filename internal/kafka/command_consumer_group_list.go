package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newGroupListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumer groups.",
		Args:  cobra.NoArgs,
		RunE:  c.groupList,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) listGroups(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	consumerGroups, err := kafkaREST.CloudClient.ListKafkaConsumerGroups()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, group := range consumerGroups.Data {
		list.Add(&consumerGroupOut{
			ClusterId:         group.GetClusterId(),
			ConsumerGroupId:   group.GetConsumerGroupId(),
			Coordinator:       getStringBroker(group.GetCoordinator()),
			IsSimple:          group.GetIsSimple(),
			PartitionAssignor: group.GetPartitionAssignor(),
			State:             group.GetState(),
		})
	}
	return list.Print()
}
