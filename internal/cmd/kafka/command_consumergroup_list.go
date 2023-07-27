package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *consumerGroupCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumer groups.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerGroupCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	consumerGroups, err := kafkaREST.CloudClient.ListKafkaConsumerGroups(cluster.ID)
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
