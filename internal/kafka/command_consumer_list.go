package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
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
	cmd.Flags().String("endpoint", "", "Endpoint to be used for this Kafka cluster.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))

	return cmd
}

func (c *consumerCommand) list(cmd *cobra.Command, _ []string) error {
	err := pcmd.SpecifyEndpoint(cmd, c.AuthenticatedCLICommand)
	if err != nil {
		return err
	}

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
	for _, consumer := range consumers {
		list.Add(&consumerOut{
			ConsumerGroup: consumer.GetConsumerGroupId(),
			Consumer:      consumer.GetConsumerId(),
			Instance:      consumer.GetInstanceId(),
			Client:        consumer.GetClientId(),
		})
	}
	return list.Print()
}
