package network

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
)

func (c *accessPointCommand) newPrivateLinkCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "private-link",
		Short: "Manage access point private links.",
	}

	cmd.AddCommand(c.newEgressEndpointCommand())
	cmd.AddCommand(c.newIngressEndpointCommand(cfg))

	return cmd
}
