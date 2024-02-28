package network

import (
	"github.com/spf13/cobra"
)

func (c *command) newDnsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS resources.",
	}

	cmd.AddCommand(c.newDnsRecordCommand())

	return cmd
}
