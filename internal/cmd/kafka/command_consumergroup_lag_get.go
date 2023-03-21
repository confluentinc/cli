package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *lagCommand) newGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "get <consumer-group>",
		Short:             "Get consumer lag for a Kafka topic partition.",
		Long:              "Get consumer lag for a Kafka topic partition consumed by a consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.getLag,
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
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("topic"))
	cobra.CheckErr(cmd.MarkFlagRequired("partition"))

	return cmd
}

func (c *lagCommand) getLag(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	partition, err := cmd.Flags().GetInt32("partition")
	if err != nil {
		return err
	}

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	lagGetResp, httpResp, err := kafkaREST.CloudClient.GetKafkaConsumerLag(lkc, consumerGroupId, topic, partition)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	table := output.NewTable(cmd)
	table.Add(convertLagToStruct(lagGetResp))
	return table.Print()
}
