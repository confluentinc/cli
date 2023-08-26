package admin

import (
	"github.com/confluentinc/cli/v3/pkg/color"
	"github.com/spf13/cobra"
)

func (c *command) newAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <code>",
		Short: "Add a new promo code.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.add,
	}
}

func (c *command) add(_ *cobra.Command, args []string) error {
	user, err := c.Client.Auth.User()
	if err != nil {
		return err
	}

	if _, err := c.Client.Billing.ClaimPromoCode(user.GetOrganization(), args[0]); err != nil {
		return err
	}

	color.Println(c.Config.EnableColor, "Your promo code was successfully added.")
	return nil
}
