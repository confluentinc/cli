package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *consumerCommand) newGroupListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumer groups.",
		Args:  cobra.NoArgs,
		RunE:  c.groupList,
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) groupList(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	groups, err := kafkaREST.CloudClient.ListKafkaConsumerGroups()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, group := range groups {
		list.Add(&consumerGroupOut{
			Cluster:           group.GetClusterId(),
			ConsumerGroup:     group.GetConsumerGroupId(),
			Coordinator:       getStringBroker(group.GetCoordinator().Related),
			IsSimple:          group.GetIsSimple(),
			PartitionAssignor: group.GetPartitionAssignor(),
			State:             group.GetState(),
		})
	}
	return list.Print()
}
