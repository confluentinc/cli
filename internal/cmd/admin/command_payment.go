package admin

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func NewPaymentCommand(prerunner pcmd.PreRunner) *cobra.Command {
	c := &command{
		pcmd.NewAuthenticatedCLICommand(
			&cobra.Command{
				Use:   "payment",
				Short: "Manage linked credit cards for an organization.",
				Args:  cobra.NoArgs,
			},
			prerunner,
		),
	}

	c.AddCommand(c.newDescribeCommand())
	c.AddCommand(c.newUpdateCommand())

	return c.Command
}

func (c *command) newDescribeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "describe",
		Short: "Print payment info for an organization.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.describe),
	}
}

func (c *command) describe(cmd *cobra.Command, _ []string) error {
	org := &orgv1.Organization{Id: c.State.Auth.User.OrganizationId}
	card, err := c.Client.Organization.GetPaymentInfo(context.Background(), org)
	if err != nil {
		return err
	}

	pcmd.Println(cmd, card)
	return nil
}

func (c *command) newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update payment info for an organization.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.update),
	}
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	// TODO: Show name of organization
	org := &orgv1.Organization{Id: c.State.Auth.User.OrganizationId}
	stripeToken := ""

	if err := c.Client.Organization.UpdatePaymentInfo(context.Background(), org, stripeToken); err != nil {
		return err
	}

	pcmd.Println(cmd, "Updated payment info.")
	return nil
}
