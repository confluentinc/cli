package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *consumerCommand) newListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumers.",
		Args:  cobra.NoArgs,
		RunE:  c.listOnPrem,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List all consumers for consumer-group "my-consumer-group".`,
				Code: "confluent kafka consumer list --consumer-group my-consumer-group",
			},
		),
	}

	cmd.Flags().String("consumer-group", "", "Consumer group ID.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("consumer-group"))

	return cmd
}

func (c *consumerCommand) listOnPrem(cmd *cobra.Command, _ []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}

	consumerGroup, err := cmd.Flags().GetString("consumer-group")
	if err != nil {
		return err
	}

	consumerDataList, resp, err := restClient.ConsumerGroupV3Api.ListKafkaConsumers(restContext, clusterId, consumerGroup)
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	list := output.NewList(cmd)
	for _, consumer := range consumerDataList.Data {
		instanceId := ""
		if consumer.InstanceId != nil {
			instanceId = *consumer.InstanceId
		}

		list.Add(&consumerOut{
			ConsumerGroupId: consumer.ConsumerGroupId,
			ConsumerId:      consumer.ConsumerId,
			InstanceId:      instanceId,
			ClientId:        consumer.ClientId,
		})
	}
	return list.Print()
}
