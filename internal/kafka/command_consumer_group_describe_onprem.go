package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newGroupDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <group>",
		Short: "Describe a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.groupDescribeOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) groupDescribeOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	group, resp, err := restClient.ConsumerGroupV3Api.GetKafkaConsumerGroup(restContext, clusterId, args[0])
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	table := output.NewTable(cmd)
	table.Add(&consumerGroupOut{
		ClusterId:         group.ClusterId,
		ConsumerGroupId:   group.ConsumerGroupId,
		Coordinator:       getStringBroker(group.Coordinator.Related),
		IsSimple:          group.IsSimple,
		PartitionAssignor: group.PartitionAssignor,
		State:             group.State,
	})
	return table.Print()
}
