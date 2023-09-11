package network

import (
	"github.com/spf13/cobra"
)

func (c *command) newPrivateLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "private-link",
		Short:   "Manage private links.",
		Aliases: []string{"pl"},
	}

	cmd.AddCommand(c.newPrivateLinkAccessCommand())

	return cmd
}
