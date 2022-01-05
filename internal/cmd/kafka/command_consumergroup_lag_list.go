package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *lagCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <consumer-group>",
		Short:             "List consumer lags for a Kafka consumer group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all consumer lags for consumers in the `my-consumer-group` consumer-group.",
				Code: "confluent kafka consumer-group lag list my-consumer-group",
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

func (c *lagCommand) list(cmd *cobra.Command, args []string) error {
	consumerGroupId := args[0]

	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	lagSummaryResp, httpResp, err := kafkaREST.Client.ConsumerGroupApi.ClustersClusterIdConsumerGroupsConsumerGroupIdLagsGet(kafkaREST.Context, lkc, consumerGroupId)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, lagFields, lagListHumanLabels, lagListStructuredLabels)
	if err != nil {
		return err
	}

	for _, lagData := range lagSummaryResp.Data {
		outputWriter.AddElement(convertLagToStruct(lagData))
	}

	return outputWriter.Out()
}
