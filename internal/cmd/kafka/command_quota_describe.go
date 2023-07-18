package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	nameconversions "github.com/confluentinc/cli/internal/pkg/name-conversions"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *quotaCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id|name>",
		Short:             "Describe a Kafka client quota.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *quotaCommand) describe(cmd *cobra.Command, args []string) error {
	quotaId := args[0]

	quota, err := c.V2Client.DescribeKafkaQuota(quotaId)
	if err != nil {
		quotaId, err = nameconversions.ConvertQuotaNameToId(quotaId, c.Context.KafkaClusterContext.GetActiveKafkaClusterId(), c.Context.GetCurrentEnvironment(), c.V2Client)
		if err != nil {
			return err
		}
		if quota, err = c.V2Client.DescribeKafkaQuota(quotaId); err != nil {
			return err
		}
	}

	table := output.NewTable(cmd)
	format := output.GetFormat(cmd)
	table.Add(quotaToPrintable(quota, format))
	return table.Print()
}
