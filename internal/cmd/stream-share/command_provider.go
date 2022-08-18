package streamshare

import (
	"github.com/spf13/cobra"
)

func (c *command) newProviderCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage provider actions.",
	}

	cmd.AddCommand(c.newProviderShareCommand())

	return cmd
}
