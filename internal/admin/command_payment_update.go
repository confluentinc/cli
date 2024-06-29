package admin

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/admin"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update the active payment method.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
	}
}

func (c *command) update(_ *cobra.Command, _ []string) error {
	output.Println(c.Config.EnableColor, "Edit credit card")

	f := form.New(
		form.Field{ID: "card number", Prompt: "Card number", Regex: `^(?:\d[ -]*?){13,19}$`},
		form.Field{ID: "expiration", Prompt: "MM/YY", Regex: `^\d{2}/\d{2}$`},
		form.Field{ID: "cvc", Prompt: "CVC", Regex: `^\d{3,4}$`, IsHidden: true},
		form.Field{ID: "name", Prompt: "Cardholder name"},
	)

	if err := f.Prompt(form.NewPrompt()); err != nil {
		return err
	}

	user, err := c.Client.Auth.User()
	if err != nil {
		return err
	}

	stripeToken, err := admin.NewStripeToken(c.Config, f.Responses["card number"].(string), f.Responses["expiration"].(string), f.Responses["cvc"].(string), f.Responses["name"].(string))
	if err != nil {
		return err
	}

	if err := c.Client.Billing.UpdatePaymentInfo(user.GetOrganization(), stripeToken.ID); err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, "Updated.")
	return nil
}
