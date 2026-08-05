package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *shareGroupCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka share groups.",
		Args:  cobra.NoArgs,
		RunE:  c.listOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *shareGroupCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	groups, resp, err := restClient.ShareGroupV3Api.ListKafkaShareGroups(restContext, clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, group := range groups.Data {
		list.Add(&shareGroupOut{
			Cluster:       group.ClusterId,
			ShareGroup:    group.ShareGroupId,
			Coordinator:   getStringBroker(group.Coordinator.Related),
			State:         group.State,
			ConsumerCount: group.ConsumerCount,
		})
	}
	list.Filter([]string{"Cluster", "ShareGroup", "Coordinator", "State", "ConsumerCount"})
	return list.Print()
}
