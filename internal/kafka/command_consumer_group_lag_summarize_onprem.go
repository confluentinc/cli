package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/kafkarest"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *consumerCommand) newLagSummarizeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "summarize <group>",
		Short:             "Summarize consumer lag for a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		RunE:              c.summarizeOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) summarizeOnPrem(cmd *cobra.Command, args []string) error {
	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
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
		ClusterId:         summary.ClusterId,
		ConsumerGroupId:   summary.ConsumerGroupId,
		TotalLag:          summary.TotalLag,
		MaxLag:            summary.MaxLag,
		MaxLagConsumerId:  summary.MaxLagConsumerId,
		MaxLagInstanceId:  maxLagInstanceId,
		MaxLagClientId:    summary.MaxLagClientId,
		MaxLagTopicName:   summary.MaxLagTopicName,
		MaxLagPartitionId: summary.MaxLagPartitionId,
	})
	return table.Print()
}
