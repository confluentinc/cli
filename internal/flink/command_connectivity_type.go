package flink

import (
	"github.com/spf13/cobra"
)

var fields = []string{"private", "public"}

func (c *command) newConnectivityTypeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connectivity-type",
		Short: "Manage Flink connectivity type.",
	}

	cmd.AddCommand(c.newUseCommand())

	return cmd
}
