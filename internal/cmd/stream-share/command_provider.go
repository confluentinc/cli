package streamshare

import (
	"github.com/spf13/cobra"
)

func (c *command) newProviderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage provider actions.",
	}

	cmd.AddCommand(c.newInviteCommand())
	cmd.AddCommand(c.newOptInCommand())
	cmd.AddCommand(c.newOptOutCommand())
	cmd.AddCommand(c.newProviderShareCommand())

	return cmd
}
