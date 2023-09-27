package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newLagGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "get <group>",
		Short:             "Get consumer lag for a Kafka topic partition.",
		Long:              "Get consumer lag for a Kafka topic partition consumed by a consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGroupArgs),
		RunE:              c.get,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Get the consumer lag for topic "my-topic" partition "0" consumed by consumer group "my-consumer-group".`,
				Code: "confluent kafka consumer group lag get my-consumer-group --topic my-topic --partition 0",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name.")
	cmd.Flags().Int32("partition", 0, "Partition ID.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("topic"))
	cobra.CheckErr(cmd.MarkFlagRequired("partition"))

	return cmd
}

func (c *consumerCommand) get(cmd *cobra.Command, args []string) error {
	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	partition, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	consumerLag, err := kafkaREST.CloudClient.GetKafkaConsumerLag(args[0], topic, partition)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&lagOut{
		ClusterId:       consumerLag.GetClusterId(),
		ConsumerGroupId: consumerLag.GetConsumerGroupId(),
		Lag:             consumerLag.GetLag(),
		LogEndOffset:    consumerLag.GetLogEndOffset(),
		CurrentOffset:   consumerLag.GetCurrentOffset(),
		ConsumerId:      consumerLag.GetConsumerId(),
		InstanceId:      consumerLag.GetInstanceId(),
		ClientId:        consumerLag.GetClientId(),
		TopicName:       consumerLag.GetTopicName(),
		PartitionId:     consumerLag.GetPartitionId(),
	})
	return table.Print()
}
