package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/spf13/cobra"
)

func (c *quotaCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a client quota.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
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
		return quotaErr(err)
	}
	printableQuota := quotaToPrintable(quota)
	return output.DescribeObject(cmd, printableQuota, quotaListFields, humanRenames, structuredRenames)
}
