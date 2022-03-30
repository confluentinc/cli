package admin

import (
	"context"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDescribeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "describe",
		Short: "Describe the active payment method.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.describe),
	}
}

func (c *command) describe(cmd *cobra.Command, _ []string) error {
	org := &orgv1.Organization{Id: c.State.Auth.Organization.Id}

	card, err := c.Client.Billing.GetPaymentInfo(context.Background(), org)
	if err != nil {
		return err
	}

	if card == nil {
		utils.Println(cmd, "Payment method not found. Add one using `confluent admin payment update`.")
		return nil
	}

	utils.Printf(cmd, "%s ending in %s\n", card.Brand, card.Last4)
	return nil
}
