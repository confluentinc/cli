package admin

import (
	"context"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

type promoCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func NewPromoCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &promoCommand{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "promo",
				Short: "Manage promo codes.",
				Args:  cobra.NoArgs,
			},
			prerunner,
		),
	}

	c.AddCommand(c.newAddCommand())

	return c.Command
}

func (c *promoCommand) newAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <code>",
		Short: "Add a new promo code.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.addRunE),
	}
}

func (c *promoCommand) addRunE(cmd *cobra.Command, args []string) error {
	org := &orgv1.Organization{Id: c.State.Auth.User.OrganizationId}
	code := args[0]

	if err := c.Client.Billing.ClaimPromoCode(context.Background(), org, code); err != nil {
		return err
	}

	utils.Println(cmd, "Your promo code was successfully added.")
	return nil
}
