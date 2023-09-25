package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumers.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List all consumers in consumer group "my-consumer-group".`,
				Code: "confluent kafka consumer list --group my-consumer-group",
			},
		),
	}

	c.addConsumerGroupFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))

	return cmd
}

func (c *consumerCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	consumers, err := kafkaREST.CloudClient.ListKafkaConsumers(group)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, consumer := range consumers.GetData() {
		list.Add(&consumerOut{
			ConsumerGroupId: consumer.GetConsumerGroupId(),
			ConsumerId:      consumer.GetConsumerId(),
			InstanceId:      consumer.GetInstanceId(),
			ClientId:        consumer.GetClientId(),
		})
	}
	return list.Print()
}
