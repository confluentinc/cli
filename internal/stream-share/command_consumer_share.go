package streamshare

import (
	"time"

	"github.com/spf13/cobra"
)

type consumerShareHumanOut struct {
	Id                         string    `human:"ID"`
	ProviderName               string    `human:"Provider Name"`
	ProviderOrganizationName   string    `human:"Provider Organization Name"`
	Status                     string    `human:"Status"`
	InviteExpiresAt            time.Time `human:"Invite Expires At"`
	NetworkDnsDomain           string    `human:"Network DNS Domain"`
	NetworkZones               string    `human:"Network Zones"`
	NetworkZonalSubdomains     string    `human:"Network Zonal Subdomains"`
	NetworkKind                string    `human:"Network Kind"`
	NetworkPrivateLinkDataType string    `human:"Network Private Link Data Type"`
	NetworkPrivateLinkData     string    `human:"Network Private Link Data"`
}

type consumerShareSerializedOut struct {
	Id                         string    `serialized:"id"`
	ProviderName               string    `serialized:"provider_name"`
	ProviderOrganizationName   string    `serialized:"provider_organization_name"`
	Status                     string    `serialized:"status"`
	InviteExpiresAt            time.Time `serialized:"invite_expires_at"`
	NetworkDnsDomain           string    `serialized:"network_dns_domain"`
	NetworkZones               string    `serialized:"network_zones"`
	NetworkZonalSubdomains     []string  `serialized:"network_zonal_subdomains"`
	NetworkKind                string    `serialized:"network_kind"`
	NetworkPrivateLinkDataType string    `serialized:"network_private_link_data_type"`
	NetworkPrivateLinkData     string    `serialized:"network_private_link_data"`
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
