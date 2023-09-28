package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newLagDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <group>",
		Short:             "Describe consumer lag for a Kafka topic partition.",
		Long:              "Describe consumer lag for a Kafka topic partition consumed by a consumer group.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.groupLagDescribeOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the consumer lag for topic `my-topic` partition `0` consumed by consumer group `my-consumer-group`.",
				Code: "confluent kafka consumer group lag describe my-consumer-group --topic my-topic --partition 0",
			},
		),
	}

	cmd.Flags().String("topic", "", "Topic name.")
	cmd.Flags().Int32("partition", 0, "Partition ID.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("topic"))
	cobra.CheckErr(cmd.MarkFlagRequired("partition"))

	return cmd
}

func (c *consumerCommand) groupLagDescribeOnPrem(cmd *cobra.Command, args []string) error {
	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	partition, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}

	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	consumerLag, resp, err := restClient.PartitionV3Api.GetKafkaConsumerLag(restContext, clusterId, args[0], topic, partition)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	table := output.NewTable(cmd)
	instanceId := ""
	if consumerLag.InstanceId != nil {
		instanceId = *consumerLag.InstanceId
	}
	table.Add(&lagOut{
		ClusterId:       consumerLag.ClusterId,
		ConsumerGroupId: consumerLag.ConsumerGroupId,
		Lag:             consumerLag.Lag,
		LogEndOffset:    consumerLag.LogEndOffset,
		CurrentOffset:   consumerLag.CurrentOffset,
		ConsumerId:      consumerLag.ConsumerId,
		InstanceId:      instanceId,
		ClientId:        consumerLag.ClientId,
		Topic:           consumerLag.TopicName,
		PartitionId:     consumerLag.PartitionId,
	})
	return table.Print()
}
