package streamshare

import (
	"strings"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
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
	shareId := args[0]

	consumerShare, err := c.V2Client.DescribeConsumerShare(args[0])
	if err != nil {
		return err
	}

	consumerSharedResources, err := c.V2Client.ListConsumerSharedResources(shareId)
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

	out := c.buildConsumerShare(consumerShare)
	cloud := network.GetCloud()
	if cloud.CdxV1AwsNetwork == nil && cloud.CdxV1AzureNetwork == nil && cloud.CdxV1GcpNetwork == nil {
		table := output.NewTable(cmd)
		table.Add(out)
		table.Filter([]string{"Id", "ProviderName", "ProviderOrganizationName", "Status", "InviteExpiresAt"})
		return table.Print()
	}

	return c.handlePrivateLinkClusterConsumerShare(cmd, network, out)
}

func (c *command) handlePrivateLinkClusterConsumerShare(cmd *cobra.Command, network cdxv1.CdxV1Network, out *consumerShareOut) error {
	networkDetails := getPrivateLinkNetworkDetails(network)
	out.NetworkDnsDomain = network.GetDnsDomain()
	out.NetworkZones = strings.Join(network.GetZones(), ",")
	out.NetworkZonalSubdomains = mapSubdomainsToList(network.GetZonalSubdomains())
	out.NetworkKind = networkDetails.networkKind
	out.NetworkPrivateLinkDataType = networkDetails.privateLinkDataType
	out.NetworkPrivateLinkData = networkDetails.privateLinkData

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
