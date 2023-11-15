package iam

import (
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

func (c *ipGroupCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe an IP group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describe,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipGroupCommand) describe(cmd *cobra.Command, args []string) error {

	group, err := c.V2Client.GetIamIpGroup(args[0])
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanIPGroup(cmd, group)
	}
	return printSerializedIPGroup(cmd, group)
}
