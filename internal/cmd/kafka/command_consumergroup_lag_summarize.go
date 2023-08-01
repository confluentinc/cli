package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
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

func (c *lagCommand) newSummarizeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "summarize <consumer-group>",
		Short:             "Summarize consumer lag for a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.summarize,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Summarize the lag for the `my-consumer-group` consumer-group.",
				Code: "confluent kafka consumer-group lag summarize my-consumer-group",
			},
		),
		Hidden: true,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *lagCommand) summarize(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(c.AuthenticatedCLICommand)
	if err != nil {
		return err
	}

	summary, err := kafkaREST.CloudClient.GetKafkaConsumerGroupLagSummary(lkc, consumerGroupId)
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
