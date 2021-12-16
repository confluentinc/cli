package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *lagCommand) newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <consumer-group>",
		Short: "Get consumer lag for a Kafka topic partition.",
		Long:  "Get consumer lag for a Kafka topic partition consumed by a consumer group.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.getLag),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Get the consumer lag for topic `my-topic` partition `0` consumed by consumer-group `my-consumer-group`.",
				Code: "confluent kafka consumer-group lag get my-consumer-group --topic my-topic --partition 0",
			},
		),
		Hidden: true,
	}

	cmd.Flags().String("topic", "", "Topic name.")
	cmd.Flags().Int32("partition", 0, "Partition ID.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("topic")
	_ = cmd.MarkFlagRequired("partition")

	return cmd
}

func (c *lagCommand) getLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	topicName, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	partitionId, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	lagGetResp, httpResp, err := kafkaREST.Client.PartitionApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagsTopicNamePartitionsPartitionIdGet(kafkaREST.Context, lkc, consumerGroupId, topicName, partitionId)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return output.DescribeObject(cmd, convertLagToStruct(lagGetResp), lagFields, lagGetHumanRenames, lagGetStructuredRenames)
}
