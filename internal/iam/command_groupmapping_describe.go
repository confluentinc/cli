package iam

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

func (c *groupMappingCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a group mapping.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *groupMappingCommand) describe(cmd *cobra.Command, args []string) error {
	groupMapping, err := c.V2Client.GetGroupMapping(args[0])
	if err != nil {
		return err
	}
	return printGroupMapping(cmd, groupMapping)
}
