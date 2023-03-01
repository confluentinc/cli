package admin

import (
	"context"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/utils"
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
	marketplace := c.Context.GetOrganization().GetMarketplace()

	card, err := c.Client.Billing.GetPaymentInfo(context.Background(), org)
	if err != nil {
		return err
	}

	if marketplace != nil && marketplace.GetPartner() != ccloudv1.MarketplacePartner_UNKNOWN {
		utils.Printf("Organization is currently linked to %s Marketplace account.\n", marketplace.GetPartner())
	}

	if card == nil {
		utils.Println("No credit card found. Add one using `confluent admin payment update`.")
		return nil
	}

	utils.Printf("%s ending in %s\n", card.GetBrand(), card.GetLast4())
	return nil
}
