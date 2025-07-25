package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type summarizeOut struct {
	Cluster         string `human:"Cluster" serialized:"cluster"`
	ConsumerGroup   string `human:"Consumer Group" serialized:"consumer_group"`
	TotalLag        int64  `human:"Total Lag" serialized:"total_lag"`
	MaxLag          int64  `human:"Max Lag" serialized:"max_lag"`
	MaxLagConsumer  string `human:"Max Lag Consumer" serialized:"max_lag_consumer"`
	MaxLagInstance  string `human:"Max Lag Instance" serialized:"max_lag_instance"`
	MaxLagClient    string `human:"Max Lag Client" serialized:"max_lag_client"`
	MaxLagTopic     string `human:"Max Lag Topic" serialized:"max_lag_topic"`
	MaxLagPartition int32  `human:"Max Lag Partition" serialized:"max_lag_partition"`
}

func (c *consumerCommand) newLagSummarizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "summarize <group>",
		Short:             "Summarize consumer lag for a Kafka consumer group.",
		Long:              "Summarize consumer lag for a Kafka consumer group. Only available for dedicated Kafka clusters.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGroupArgs),
		RunE:              c.groupLagSummarize,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) groupLagSummarize(cmd *cobra.Command, args []string) error {
	if err := c.checkIsDedicated(); err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	summary, err := kafkaREST.CloudClient.GetKafkaConsumerGroupLagSummary(args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&summarizeOut{
		Cluster:         summary.GetClusterId(),
		ConsumerGroup:   summary.GetConsumerGroupId(),
		TotalLag:        summary.GetTotalLag(),
		MaxLag:          summary.GetMaxLag(),
		MaxLagConsumer:  summary.GetMaxLagConsumerId(),
		MaxLagInstance:  summary.GetMaxLagInstanceId(),
		MaxLagClient:    summary.GetMaxLagClientId(),
		MaxLagTopic:     summary.GetMaxLagTopicName(),
		MaxLagPartition: summary.GetMaxLagPartitionId(),
	})
	return table.Print()
}
