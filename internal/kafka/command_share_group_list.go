package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *shareCommand) newGroupListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka share groups.",
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

func (c *shareCommand) groupList(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	shareGroups, err := kafkaREST.CloudClient.ListKafkaShareGroups()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, shareGroup := range shareGroups {
		list.Add(&shareGroupListOut{
			Cluster:       shareGroup.GetClusterId(),
			ShareGroup:    shareGroup.GetShareGroupId(),
			Coordinator:   getStringBroker(shareGroup.GetCoordinator().Related),
			State:         shareGroup.GetState(),
			ConsumerCount: shareGroup.GetConsumerCount(),
		})
	}
	return list.Print()
}
