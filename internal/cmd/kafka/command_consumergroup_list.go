package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	groupListFields           = []string{"ClusterId", "ConsumerGroupId", "IsSimple", "State"}
	groupListHumanLabels      = []string{"Cluster", "Consumer Group", "Simple", "State"}
	groupListStructuredLabels = []string{"cluster", "consumer_group", "simple", "state"}
)

func (c *consumerGroupCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka consumer groups.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all consumer groups.",
				Code: "confluent kafka consumer-group list",
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

func (c *consumerGroupCommand) list(cmd *cobra.Command, _ []string) error {
	kafkaREST, lkc, err := getKafkaRestProxyAndLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	groupCmdResp, httpResp, err := kafkaREST.Client.ConsumerGroupV3Api.ListKafkaConsumerGroups(kafkaREST.Context, lkc).Execute()
	if err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, groupListFields, groupListHumanLabels, groupListStructuredLabels)
	if err != nil {
		return err
	}

	for _, groupData := range groupCmdResp.Data {
		outputWriter.AddElement(&groupData)
	}

	return outputWriter.Out()
}
