package streamshare

import (
	"time"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
)

type consumerShareOut struct {
	Id                         string    `human:"ID" serialized:"id"`
	ProviderName               string    `human:"Provider Name" serialized:"provider_name"`
	ProviderOrganizationName   string    `human:"Provider Organization Name" serialized:"provider_organization_name"`
	Status                     string    `human:"Status" serialized:"status"`
	InviteExpiresAt            time.Time `human:"Invite Expires At" serialized:"invite_expires_at"`
	NetworkDnsDomain           string    `human:"Network DNS Domain" serialized:"network_dns_domain"`
	NetworkZones               string    `human:"Network Zones" serialized:"network_zones"`
	NetworkZonalSubdomains     []string  `human:"Network Zonal Subdomains" serialized:"network_zonal_subdomains"`
	NetworkKind                string    `human:"Network Kind" serialized:"network_kind"`
	NetworkPrivateLinkDataType string    `human:"Network Private Link Data Type" serialized:"network_private_link_data_type"`
	NetworkPrivateLinkData     string    `human:"Network Private Link Data" serialized:"network_private_link_data"`
}

func (c *command) newConsumerShareCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "share",
		Short: "Manage consumer shares.",
	}

	cmd.AddCommand(c.newConsumerShareDeleteCommand())
	cmd.AddCommand(c.newConsumerShareDescribeCommand())
	cmd.AddCommand(c.newConsumerShareListCommand())

	return cmd
}

func (c *command) validConsumerShareArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validConsumerShareArgsMultiple(cmd, args)
}

func (c *command) validConsumerShareArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteConsumerShares()
}

func (c *command) autocompleteConsumerShares() []string {
	consumerShares, err := c.V2Client.ListConsumerShares("")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(consumerShares))
	for i, share := range consumerShares {
		suggestions[i] = *share.Id
	}
	return suggestions
}

func (c *command) buildConsumerShare(share cdxv1.CdxV1ConsumerShare) *consumerShareOut {
	status := share.GetStatus()
	return &consumerShareOut{
		Id:                       share.GetId(),
		ProviderName:             share.GetProviderUserName(),
		ProviderOrganizationName: share.GetProviderOrganizationName(),
		Status:                   status.GetPhase(),
		InviteExpiresAt:          share.GetInviteExpiresAt(),
	}
}
