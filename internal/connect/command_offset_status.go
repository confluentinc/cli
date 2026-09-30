package connect

import (
	"github.com/spf13/cobra"
)

func (c *offsetCommand) newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Manage the status of an offset update or delete.",
	}

	cmd.AddCommand(c.newStatusDescribeCommand())

	return cmd
}
