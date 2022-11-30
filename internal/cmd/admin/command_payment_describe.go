package admin

import (
	"context"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

func (c *command) newDescribeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "describe",
		Short: "Describe the active payment method.",
		Args:  cobra.NoArgs,
		RunE:  c.describe,
	}
}

func (c *command) describe(cmd *cobra.Command, _ []string) error {
	org := &ccloudv1.Organization{Id: c.Context.GetOrganization().GetId()}

	card, err := c.Client.Billing.GetPaymentInfo(context.Background(), org)
	if err != nil {
		return err
	}

	if card == nil {
		utils.Println(cmd, "Payment method not found. Add one using `confluent admin payment update`.")
		return nil
	}

	utils.Printf(cmd, "%s ending in %s\n", card.GetBrand(), card.GetLast4())
	return nil
}
