package streamshare

import (
	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newConsumerShareDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a consumer share.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validConsumerShareArgs),
		RunE:              c.describeConsumerShare,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe consumer share "ss-12345":`,
				Code: "confluent stream-share consumer share describe ss-12345",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describeConsumerShare(cmd *cobra.Command, args []string) error {
	share, err := c.V2Client.DescribeConsumerShare(args[0])
	if err != nil {
		return err
	}

	consumerSharedResources, err := c.V2Client.ListConsumerSharedResources(args[0])
	if err != nil {
		return err
	}

	var network cdxv1.CdxV1Network
	if len(consumerSharedResources) > 0 {
		privateNetwork, err := c.V2Client.GetPrivateLinkNetworkConfig(consumerSharedResources[0].GetId())
		if err != nil {
			return err
		}
		network = privateNetwork
	}

	cloud := network.GetCloud()
	isPrivateLink := cloud.CdxV1AwsNetwork != nil || cloud.CdxV1AzureNetwork != nil || cloud.CdxV1GcpNetwork != nil

	table := output.NewTable(cmd)
	if isPrivateLink {
		networkDetails := getPrivateLinkNetworkDetails(network)

		table.Add(&consumerShareOut{
			Id:                         share.GetId(),
			ProviderName:               share.GetProviderUserName(),
			ProviderOrganizationName:   share.GetProviderOrganizationName(),
			Status:                     share.Status.GetPhase(),
			InviteExpiresAt:            share.GetInviteExpiresAt(),
			NetworkDnsDomain:           network.GetDnsDomain(),
			NetworkZones:               network.GetZones(),
			NetworkZonalSubdomains:     network.GetZonalSubdomains(),
			NetworkKind:                networkDetails.networkKind,
			NetworkPrivateLinkDataType: networkDetails.privateLinkDataType,
			NetworkPrivateLinkData:     networkDetails.privateLinkData,
		})
	} else {
		table.Add(&consumerShareOut{
			Id:                       share.GetId(),
			ProviderName:             share.GetProviderUserName(),
			ProviderOrganizationName: share.GetProviderOrganizationName(),
			Status:                   share.Status.GetPhase(),
			InviteExpiresAt:          share.GetInviteExpiresAt(),
		})

		table.Filter([]string{"Id", "ProviderName", "ProviderOrganizationName", "Status", "InviteExpiresAt"})
	}

	return table.Print()
}
