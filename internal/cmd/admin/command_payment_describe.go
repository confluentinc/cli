package admin

import (
	"context"
	"fmt"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/featureflags"
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
		utils.Println(cmd, fmt.Sprintf("Organization is currently linked to %s Marketplace account.", marketplace.GetPartner()))
	}

	if card == nil {
		utils.Println(cmd, "No credit card found. Add one using `confluent admin payment update`.")

		ldClient, err := v1.GetCcloudLaunchDarklyClient(c.Context.PlatformName)
		if err != nil {
			log.CliLogger.Debugf("Skip conditionally advertising Marketplace payment option due to error: %s", err.Error())
			return nil
		}

		// if experiment for advertising Marketplace payment option is enabled, then add a copy
		if featureflags.Manager.BoolVariation("cloud_growth.marketplace_linking_advertisement_experiment.enable", c.Context, ldClient, true, false) {
			utils.Println(cmd, "Alternatively, you can also link to AWS, GCP, or Azure Marketplace as your payment option. For more information, visit https://confluent.cloud/add-payment.")
		}

		return nil
	}

	utils.Printf(cmd, "%s ending in %s\n", card.GetBrand(), card.GetLast4())
	return nil
}
