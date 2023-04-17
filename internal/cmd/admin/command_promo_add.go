package admin

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/output"
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
	user, err := c.Client.Auth.User(context.Background())
	if err != nil {
		return err
	}

	if _, err := c.Client.Billing.ClaimPromoCode(context.Background(), user.GetOrganization(), args[0]); err != nil {
		return err
	}

	output.Println("Your promo code was successfully added.")
	return nil
}
