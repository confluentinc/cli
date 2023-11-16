package iam

import (
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

func (c *ipFilterCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe an IP filter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) describe(cmd *cobra.Command, args []string) error {

	filter, err := c.V2Client.GetIamIpFilter(args[0])
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanIpFilter(cmd, filter)
	}
	return printSerializedIpFilter(cmd, filter)
}
