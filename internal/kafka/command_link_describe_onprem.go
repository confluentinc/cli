package kafka

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *linkCommand) newDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <link>",
		Short:             "Describe a cluster link.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describeOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *linkCommand) describeOnPrem(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	client, ctx, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	linkOpts := kafkarestv3.GetKafkaLinkOpts{IncludeTasks: optional.NewBool(false)}
	data, httpResp, err := client.ClusterLinkingV3Api.GetKafkaLink(ctx, clusterId, linkName, &linkOpts)
	if err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	table := output.NewTable(cmd)
	table.Add(newLinkOnPrem(data, ""))
	table.Filter(getListFieldsOnPrem(false))
	return table.Print()
}
