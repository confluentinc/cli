package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *consumerCommand) newLagDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <group>",
		Short:             "Describe consumer lag for a Kafka topic partition.",
		Long:              "Describe consumer lag for a Kafka topic partition consumed by a consumer group. Only available for dedicated Kafka clusters.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGroupArgs),
		RunE:              c.groupLagDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the consumer lag for topic "my-topic" partition "0" consumed by consumer group "my-consumer-group".`,
				Code: "confluent kafka consumer group lag describe my-consumer-group --topic my-topic --partition 0",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name.")
	cmd.Flags().Int32("partition", 0, "Partition ID.")
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("topic"))
	cobra.CheckErr(cmd.MarkFlagRequired("partition"))

	return cmd
}

func (c *consumerCommand) groupLagDescribe(cmd *cobra.Command, args []string) error {
	if err := c.checkIsDedicated(); err != nil {
		return err
	}

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	partition, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	consumerLag, err := kafkaREST.CloudClient.GetKafkaConsumerLag(args[0], topic, partition)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&lagOut{
		Cluster:       consumerLag.GetClusterId(),
		ConsumerGroup: consumerLag.GetConsumerGroupId(),
		Lag:           consumerLag.GetLag(),
		LogEndOffset:  consumerLag.GetLogEndOffset(),
		CurrentOffset: consumerLag.GetCurrentOffset(),
		Consumer:      consumerLag.GetConsumerId(),
		Instance:      consumerLag.GetInstanceId(),
		Client:        consumerLag.GetClientId(),
		Topic:         consumerLag.GetTopicName(),
		Partition:     consumerLag.GetPartitionId(),
	})
	return table.Print()
}
