package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumers.",
		Args:  cobra.NoArgs,
		RunE:  c.listOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List all consumers in consumer group "my-consumer-group".`,
				Code: "confluent kafka consumer list --group my-consumer-group",
			},
		),
	}

	cmd.Flags().String("group", "", "Consumer group ID.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))

	return cmd
}

func (c *consumerCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	consumers, resp, err := restClient.ConsumerGroupV3Api.ListKafkaConsumers(restContext, clusterId, group)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, consumer := range consumers.Data {
		out := &consumerOut{
			ConsumerGroupId: consumer.ConsumerGroupId,
			ConsumerId:      consumer.ConsumerId,
			ClientId:        consumer.ClientId,
		}
		if consumer.InstanceId != nil {
			out.InstanceId = *consumer.InstanceId
		}
		list.Add(out)
	}
	return list.Print()
}
