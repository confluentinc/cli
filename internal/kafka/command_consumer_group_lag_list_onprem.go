package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *consumerCommand) newLagListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <group>",
		Short: "List consumer lags for a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.groupLagListOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List consumer lags in consumer group "my-consumer-group".`,
				Code: "confluent kafka consumer group lag list my-consumer-group",
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) groupLagListOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	consumerLags, resp, err := restClient.ConsumerGroupV3Api.ListKafkaConsumerLags(restContext, clusterId, args[0])
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, consumerLag := range consumerLags.Data {
		out := &lagOut{
			Cluster:       consumerLag.ClusterId,
			ConsumerGroup: consumerLag.ConsumerGroupId,
			Lag:           consumerLag.Lag,
			LogEndOffset:  consumerLag.LogEndOffset,
			CurrentOffset: consumerLag.CurrentOffset,
			Consumer:      consumerLag.ConsumerId,
			Client:        consumerLag.ClientId,
			Topic:         consumerLag.TopicName,
			Partition:     consumerLag.PartitionId,
		}
		if consumerLag.InstanceId != nil {
			out.Instance = *consumerLag.InstanceId
		}
		list.Add(out)
	}
	return list.Print()
}
