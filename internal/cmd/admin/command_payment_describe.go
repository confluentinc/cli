package admin

import (
	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
	"github.com/confluentinc/cli/internal/pkg/output"
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
	user, err := c.Client.Auth.User()
	if err != nil {
		return err
	}

	card, err := c.Client.Billing.GetPaymentInfo(user.GetOrganization())
	if err != nil {
		return err
	}

	marketplace := user.GetOrganization().GetMarketplace()
	if marketplace.GetPartner() != ccloudv1.MarketplacePartner_UNKNOWN {
		output.Printf("Organization is currently linked to %s Marketplace account.\n", marketplace.GetPartner())
	}

	if card == nil {
		output.Println("No credit card found. Add one using `confluent admin payment update`.")

		ldClient := v1.GetCcloudLaunchDarklyClient(c.Context.PlatformName)
		if featureflags.Manager.BoolVariation("cloud_growth.marketplace_linking_advertisement_experiment.enable", c.Context, ldClient, true, false) {
			output.Println("Alternatively, you can also link to AWS, GCP, or Azure Marketplace as your payment option. For more information, visit https://confluent.cloud/add-payment.")
		}

		return nil
	}

	output.Printf("%s ending in %s\n", card.GetBrand(), card.GetLast4())
	return nil
}
