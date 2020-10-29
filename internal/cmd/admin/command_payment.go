package admin

import (
	"context"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/token"
	"os"
	"strings"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	keys "github.com/confluentinc/cli/internal/pkg/third-party-keys"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
	isTest bool
}

func NewPaymentCommand(prerunner pcmd.PreRunner, isTest bool) *cobra.Command {
	c := &command{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "payment",
				Short: "Manage payment method.",
				Args:  cobra.NoArgs,
			},
			prerunner,
		),
		isTest,
	}

	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}

func (c *command) newDescribeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "describe",
		Short: "Describe the active payment method.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.describeRunE),
	}
}

func (c *command) describeRunE(cmd *cobra.Command, _ []string) error {
	org := &orgv1.Organization{Id: c.State.Auth.User.OrganizationId}
	card, err := c.Client.Organization.GetPaymentInfo(context.Background(), org)
	if err != nil {
		return err
	}
	if card == nil {
		utils.Println(cmd, "Payment method not found. Add one using \"ccloud admin payment update\".")
		return nil
	}
	utils.Printf(cmd, "%s ending in %s\n", card.Brand, card.Last4)
	return nil
}

func (c *command) newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update the active payment method.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.updateRunE),
	}
}

func (c *command) updateRunE(cmd *cobra.Command, _ []string) error {
	return c.update(cmd, form.NewPrompt(os.Stdin))
}

func (c *command) update(cmd *cobra.Command, prompt form.Prompt) error {
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

	org := &orgv1.Organization{Id: c.State.Auth.User.OrganizationId}
	if c.isTest {
		stripe.Key = keys.StripeTestKey
	} else {
		stripe.Key = keys.StripeLiveKey
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

	if err := c.Client.Organization.UpdatePaymentInfo(context.Background(), org, stripeToken.ID); err != nil {
		return err
	}

	utils.Println(cmd, "Updated.")
	return nil
}
