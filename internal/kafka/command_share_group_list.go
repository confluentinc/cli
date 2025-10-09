package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *shareGroupCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka share groups.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *shareGroupCommand) list(cmd *cobra.Command, _ []string) error {
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
		list.Add(&shareGroupOut{
			Cluster:       shareGroup.GetClusterId(),
			ShareGroup:    shareGroup.GetShareGroupId(),
			Coordinator:   getStringBroker(shareGroup.GetCoordinator().Related),
			State:         shareGroup.GetState(),
			ConsumerCount: shareGroup.GetConsumerCount(),
		})
	}
	list.Filter([]string{"Cluster", "ShareGroup", "Coordinator", "State", "ConsumerCount"})
	return list.Print()
}
