package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type summarizeOut struct {
	ClusterId         string `human:"Cluster" serialized:"cluster_id"`
	ConsumerGroupId   string `human:"Consumer Group" serialized:"consumer_group"`
	TotalLag          int64  `human:"Total Lag" serialized:"total_lag"`
	MaxLag            int64  `human:"Max Lag" serialized:"max_lag"`
	MaxLagConsumerId  string `human:"Max Lag Consumer" serialized:"max_lag_consumer"`
	MaxLagInstanceId  string `human:"Max Lag Instance" serialized:"max_lag_instance"`
	MaxLagClientId    string `human:"Max Lag Client" serialized:"max_lag_client"`
	MaxLagTopicName   string `human:"Max Lag Topic" serialized:"max_lag_topic"`
	MaxLagPartitionId int32  `human:"Max Lag Partition" serialized:"max_lag_partition"`
}

func (c *consumerCommand) newLagSummarizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "summarize <group>",
		Short:             "Summarize consumer lag for a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGroupArgs),
		RunE:              c.summarize,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) summarize(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	summary, err := kafkaREST.CloudClient.GetKafkaConsumerGroupLagSummary(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&summarizeOut{
		ClusterId:         summary.GetClusterId(),
		ConsumerGroupId:   summary.GetConsumerGroupId(),
		TotalLag:          summary.GetTotalLag(),
		MaxLag:            summary.GetMaxLag(),
		MaxLagConsumerId:  summary.GetMaxLagConsumerId(),
		MaxLagInstanceId:  summary.GetMaxLagInstanceId(),
		MaxLagClientId:    summary.GetMaxLagClientId(),
		MaxLagTopicName:   summary.GetMaxLagTopicName(),
		MaxLagPartitionId: summary.GetMaxLagPartitionId(),
	})
	return table.Print()
}
