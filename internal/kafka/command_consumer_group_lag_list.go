package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *consumerCommand) newLagListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <group>",
		Short:             "List consumer lags for a Kafka consumer group.",
		Long:              "List consumer lags for a Kafka consumer group. Only available for dedicated Kafka clusters.",
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

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) groupLagList(cmd *cobra.Command, args []string) error {
	if err := c.checkIsDedicated(); err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	consumerLags, err := kafkaREST.CloudClient.ListKafkaConsumerLags(args[0])
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, consumerLag := range consumerLags {
		list.Add(&lagOut{
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
	}
	return list.Print()
}
