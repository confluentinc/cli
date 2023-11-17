package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

func (c *ipFilterCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe an IP filter.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}
func (c *ipFilterCommand) describe(cmd *cobra.Command, args []string) error {
	filter, err := c.V2Client.GetIamIpFilter(args[0])
	if err != nil {
		return err
	}

	return printIpFilter(cmd, filter)
}
