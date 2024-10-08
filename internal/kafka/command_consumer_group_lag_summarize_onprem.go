package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/kafkarest"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *consumerCommand) newLagSummarizeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summarize <group>",
		Short: "Summarize consumer lag for a Kafka consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.groupLagSummarizeOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) groupLagSummarizeOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	summary, resp, err := restClient.ConsumerGroupV3Api.GetKafkaConsumerGroupLagSummary(restContext, clusterId, args[0])
	if err != nil {
		return kafkarest.NewError(restClient.GetConfig().BasePath, err, resp)
	}

	table := output.NewTable(cmd)
	maxLagInstanceId := ""
	if summary.MaxLagInstanceId != nil {
		maxLagInstanceId = *summary.MaxLagInstanceId
	}
	table.Add(&summarizeOut{
		Cluster:         summary.ClusterId,
		ConsumerGroup:   summary.ConsumerGroupId,
		TotalLag:        summary.TotalLag,
		MaxLag:          summary.MaxLag,
		MaxLagConsumer:  summary.MaxLagConsumerId,
		MaxLagInstance:  maxLagInstanceId,
		MaxLagClient:    summary.MaxLagClientId,
		MaxLagTopic:     summary.MaxLagTopicName,
		MaxLagPartition: summary.MaxLagPartitionId,
	})
	return table.Print()
}
