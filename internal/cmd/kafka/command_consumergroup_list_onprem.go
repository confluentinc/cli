package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *consumerGroupCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumer groups.",
		Args:  cobra.NoArgs,
		RunE:  c.listOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerGroupCommand) listOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	dataList, resp, err := restClient.ConsumerGroupV3Api.ListKafkaConsumerGroups(restContext, clusterId)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, data := range dataList.Data {
		list.Add(&consumerGroupOut{
			ClusterId:         data.ClusterId,
			ConsumerGroupId:   data.ConsumerGroupId,
			Coordinator:       getStringBrokerOnPrem(data.Coordinator),
			IsSimple:          data.IsSimple,
			PartitionAssignor: data.PartitionAssignor,
			State:             data.State,
		})
	}
	return list.Print()
}
