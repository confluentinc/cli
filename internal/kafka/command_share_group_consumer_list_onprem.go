package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *shareGroupCommand) newConsumerListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka share group consumers.",
		Args:  cobra.NoArgs,
		RunE:  c.consumerListOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List all consumers in share group "my-share-group".`,
				Code: "confluent kafka share-group consumer list --group my-share-group",
			},
		),
	}

	cmd.Flags().String("group", "", "Share group ID.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))

	return cmd
}

func (c *shareGroupCommand) consumerListOnPrem(cmd *cobra.Command, _ []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	consumers, resp, err := restClient.ShareGroupV3Api.ListKafkaShareGroupConsumers(restContext, clusterId, group)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, consumer := range consumers.Data {
		list.Add(&shareGroupConsumerOut{
			Cluster:    consumer.ClusterId,
			ShareGroup: group,
			Consumer:   consumer.ConsumerId,
			Client:     consumer.ClientId,
		})
	}
	return list.Print()
}
