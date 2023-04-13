package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *consumerGroupCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumer groups.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all consumer groups.",
				Code: "confluent kafka consumer-group list",
			},
		),
		Hidden: true,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerGroupCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(c.AuthenticatedCLICommand)
	if err != nil {
		return err
	}

	groupCmdResp, httpResp, err := kafkaREST.CloudClient.ListKafkaConsumerGroups(lkc)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	list := output.NewList(cmd)
	for _, group := range groupCmdResp.Data {
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
