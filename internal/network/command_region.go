package network

import (
	"github.com/spf13/cobra"
)

func (c *command) newRegionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "region",
		Short: "Manage Confluent Cloud network regions.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newRegionListCommand())

	return cmd
}
