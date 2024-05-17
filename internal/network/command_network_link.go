package network

import (
	"github.com/spf13/cobra"
)

func (c *command) newNetworkLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "Manage network links.",
	}

	cmd.AddCommand(c.newNetworkLinkEndpointCommand())
	cmd.AddCommand(c.newNetworkLinkServiceCommand())

	return cmd
}
