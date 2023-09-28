package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newLagListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <group>",
		Short:             "List consumer lags for a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGroupArgs),
		RunE:              c.groupLagList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List consumer lags in consumer group "my-consumer-group".`,
				Code: "confluent kafka consumer group lag list my-consumer-group",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) groupLagList(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	consumerLags, err := kafkaREST.CloudClient.ListKafkaConsumerLags(args[0])
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, consumerLag := range consumerLags.GetData() {
		list.Add(&lagOut{
			ClusterId:       consumerLag.GetClusterId(),
			ConsumerGroupId: consumerLag.GetConsumerGroupId(),
			Lag:             consumerLag.GetLag(),
			LogEndOffset:    consumerLag.GetLogEndOffset(),
			CurrentOffset:   consumerLag.GetCurrentOffset(),
			ConsumerId:      consumerLag.GetConsumerId(),
			InstanceId:      consumerLag.GetInstanceId(),
			ClientId:        consumerLag.GetClientId(),
			Topic:           consumerLag.GetTopicName(),
			PartitionId:     consumerLag.GetPartitionId(),
		})
	}
	return list.Print()
}
