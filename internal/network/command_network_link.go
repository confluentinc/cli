package network

import (
	"github.com/spf13/cobra"
)

func (c *command) newNetworkLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "network-link",
		Short:   "Manage network links.",
		Aliases: []string{"nl"},
	}

	cmd.AddCommand(c.newNetworkLinkServiceCommand())

	return cmd
}
