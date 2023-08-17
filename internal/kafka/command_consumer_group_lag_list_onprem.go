package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newLagListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <group>",
		Short:             "List consumer lags for a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.listLagsOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all consumer lags for consumers in the `my-consumer-group` consumer group.",
				Code: "confluent kafka consumer group lag list my-consumer-group",
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) listLagsOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	consumerLags, resp, err := restClient.ConsumerGroupV3Api.ListKafkaConsumerLags(restContext, clusterId, args[0])
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, consumerLag := range consumerLags.Data {
		instanceId := ""
		if consumerLag.InstanceId != nil {
			instanceId = *consumerLag.InstanceId
		}
		list.Add(&lagOut{
			ClusterId:       consumerLag.ClusterId,
			ConsumerGroupId: consumerLag.ConsumerGroupId,
			Lag:             consumerLag.Lag,
			LogEndOffset:    consumerLag.LogEndOffset,
			CurrentOffset:   consumerLag.CurrentOffset,
			ConsumerId:      consumerLag.ConsumerId,
			InstanceId:      instanceId,
			ClientId:        consumerLag.ClientId,
			TopicName:       consumerLag.TopicName,
			PartitionId:     consumerLag.PartitionId,
		})
	}
	return list.Print()
}
