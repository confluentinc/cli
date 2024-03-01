package network

import (
	"github.com/spf13/cobra"
)

func (c *accessPointCommand) newPrivateLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "private-link",
		Short: "Manage private links.",
	}

	cmd.AddCommand(c.newEgressEndpointCommand())

	return cmd
}
