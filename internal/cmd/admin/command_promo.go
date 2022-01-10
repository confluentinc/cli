package admin

import (
	"github.com/spf13/cobra"
)

func (c *command) newPromoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "promo",
		Short: "Manage promo codes.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newAddCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}
