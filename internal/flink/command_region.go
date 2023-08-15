package flink

import (
	"github.com/spf13/cobra"
)

func (c *command) newRegionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "region",
		Short: "List Flink regions.",
	}

	cmd.AddCommand(c.newRegionListCommand())
	cmd.AddCommand(c.newRegionUseCommand())

	return cmd
}
