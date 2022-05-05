package admin

import (
	"context"
	"os"
	"strings"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/token"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	stripeTestKey = "pk_test_0MJU6ihIFpxuWMwG6HhjGQ8P"
	stripeLiveKey = "pk_live_t0P8AKi9DEuvAqfKotiX5xHM"
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
	return c.updateWithPrompt(cmd, form.NewPrompt(os.Stdin))
}

func (c *command) updateWithPrompt(cmd *cobra.Command, prompt form.Prompt) error {
	utils.Println(cmd, "Edit credit card")

	f := form.New(
		form.Field{ID: "card number", Prompt: "Card number", Regex: `^(?:\d[ -]*?){13,19}$`},
		form.Field{ID: "expiration", Prompt: "MM/YY", Regex: `^\d{2}/\d{2}$`},
		form.Field{ID: "cvc", Prompt: "CVC", Regex: `^\d{3,4}$`, IsHidden: true},
		form.Field{ID: "name", Prompt: "Cardholder name"},
	)

	if err := f.Prompt(cmd, prompt); err != nil {
		return err
	}

	org := &orgv1.Organization{Id: c.State.Auth.Organization.Id}
	if c.isTest {
		stripe.Key = stripeTestKey
	} else {
		stripe.Key = stripeLiveKey
	}
	stripe.DefaultLeveledLogger = &stripe.LeveledLogger{
		Level: 0,
	}

	exp := strings.Split(f.Responses["expiration"].(string), "/")

	params := &stripe.TokenParams{
		Card: &stripe.CardParams{
			Number:   stripe.String(f.Responses["card number"].(string)),
			ExpMonth: stripe.String(exp[0]),
			ExpYear:  stripe.String(exp[1]),
			CVC:      stripe.String(f.Responses["cvc"].(string)),
			Name:     stripe.String(f.Responses["name"].(string)),
		},
	}

	stripeToken, err := token.New(params)
	if err != nil {
		if stripeErr, ok := err.(*stripe.Error); ok {
			return errors.New(stripeErr.Msg)
		}
		return err
	}

	if err := c.Client.Billing.UpdatePaymentInfo(context.Background(), org, stripeToken.ID); err != nil {
		return err
	}

	utils.Println(cmd, "Updated.")
	return nil
}
