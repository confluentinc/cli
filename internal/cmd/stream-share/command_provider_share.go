package streamshare

import (
	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
)

func (c *command) newProviderShareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Manage provider shares.",
	}

	cmd.AddCommand(c.newProviderShareDeleteCommand())
	cmd.AddCommand(c.newProviderShareDescribeCommand())
	cmd.AddCommand(c.newProviderShareListCommand())

	return cmd
}

func (c *command) validProviderShareArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteProviderShares()
}

func (c *command) autocompleteProviderShares() []string {
	providerShares, err := c.V2Client.ListProviderShares("")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(providerShares))
	for i, share := range providerShares {
		suggestions[i] = *share.Id
	}
	return suggestions
}

func (c *command) buildProviderShare(share cdxv1.CdxV1ProviderShare) *providerShare {
	serviceAccount := share.GetServiceAccount()
	sharedResource := share.GetSharedResource()
	element := &providerShare{
		Id:                       share.GetId(),
		ConsumerName:             share.GetConsumerUserName(),
		ConsumerOrganizationName: share.GetConsumerOrganizationName(),
		ProviderName:             share.GetProviderUserName(),
		Status:                   share.GetStatus(),
		DeliveryMethod:           share.GetDeliveryMethod(),
		ServiceAccountId:         serviceAccount.GetId(),
		SharedResourceId:         sharedResource.GetId(),
		InvitedAt:                share.GetInvitedAt(),
		InviteExpiresAt:          share.GetInviteExpiresAt(),
	}

	if val, ok := share.GetRedeemedAtOk(); ok && !val.IsZero() {
		element.RedeemedAt = val.String()
	}
	return element
}
