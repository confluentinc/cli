package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *consumerCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumers.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List all consumers for consumer-group "my-consumer-group".`,
				Code: "confluent kafka consumer list --consumer-group my-consumer-group",
			},
		),
	}

	c.addConsumerGroupFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("consumer-group"))

	return cmd
}

func (c *consumerCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	consumerGroup, err := cmd.Flags().GetString("consumer-group")
	if err != nil {
		return err
	}

	consumerDataList, err := kafkaREST.CloudClient.ListKafkaConsumers(cluster.ID, consumerGroup)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, consumer := range consumerDataList.Data {
		list.Add(&consumerOut{
			ConsumerGroupId: consumer.GetConsumerGroupId(),
			ConsumerId:      consumer.GetConsumerId(),
			InstanceId:      consumer.GetInstanceId(),
			ClientId:        consumer.GetClientId(),
		})
	}
	return list.Print()
}
