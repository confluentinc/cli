package admin

import (
	"context"
	"os"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

func (c *command) newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update the active payment method.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
	}
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	prompt := form.NewPrompt(os.Stdin)

	utils.Println("Edit credit card")

	f := form.New(
		form.Field{ID: "card number", Prompt: "Card number", Regex: `^(?:\d[ -]*?){13,19}$`},
		form.Field{ID: "expiration", Prompt: "MM/YY", Regex: `^\d{2}/\d{2}$`},
		form.Field{ID: "cvc", Prompt: "CVC", Regex: `^\d{3,4}$`, IsHidden: true},
		form.Field{ID: "name", Prompt: "Cardholder name"},
	)

	if err := f.Prompt(prompt); err != nil {
		return err
	}

	stripeToken, err := utils.NewStripeToken(f.Responses["card number"].(string), f.Responses["expiration"].(string), f.Responses["cvc"].(string), f.Responses["name"].(string), c.isTest)
	if err != nil {
		return err
	}

	org := &ccloudv1.Organization{Id: c.Context.GetOrganization().GetId()}
	if err := c.Client.Billing.UpdatePaymentInfo(context.Background(), org, stripeToken.ID); err != nil {
		return err
	}

	utils.Println("Updated.")
	return nil
}
