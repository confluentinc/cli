package streamshare

import (
	"time"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
)

type providerShareOut struct {
	Id                       string    `human:"ID" serialized:"id"`
	ConsumerName             string    `human:"Consumer Name" serialized:"consumer_name"`
	ConsumerOrganizationName string    `human:"Consumer Organization Name" serialized:"consumer_organization_name"`
	Status                   string    `human:"Status" serialized:"status"`
	DeliveryMethod           string    `human:"Delivery Method" serialized:"delivery_method"`
	RedeemedAt               string    `human:"Redeemed At" serialized:"redeemed_at"`
	InvitedAt                time.Time `human:"Invited At" serialized:"invited_at"`
	InviteExpiresAt          time.Time `human:"Invite Expires At" serialized:"invite_expires_at"`
}

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

func (c *command) buildProviderShare(share cdxv1.CdxV1ProviderShare) *providerShareOut {
	status := share.GetStatus()
	out := &providerShareOut{
		Id:                       share.GetId(),
		ConsumerName:             share.GetConsumerUserName(),
		ConsumerOrganizationName: share.GetConsumerOrganizationName(),
		Status:                   status.GetPhase(),
		DeliveryMethod:           share.GetDeliveryMethod(),
		InvitedAt:                share.GetInvitedAt(),
		InviteExpiresAt:          share.GetInviteExpiresAt(),
	}

	if val, ok := share.GetRedeemedAtOk(); ok && !val.IsZero() {
		out.RedeemedAt = val.String()
	}
	return out
}
