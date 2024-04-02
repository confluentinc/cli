package connect

import (
	"github.com/spf13/cobra"
)

func (c *offsetCommand) newStatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of offset update.",
	}

	cmd.AddCommand(c.newStatusDescribeCommand())
	return cmd
}
