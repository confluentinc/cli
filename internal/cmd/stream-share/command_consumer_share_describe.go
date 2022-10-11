package streamshare

import (
	"strings"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
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

	consumerShare, httpResp, err := c.V2Client.DescribeConsumerShare(shareId)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	consumerSharedResources, err := c.V2Client.ListConsumerSharedResources(shareId)
	if err != nil {
		return err
	}

	var network cdxv1.CdxV1Network
	if len(consumerSharedResources) != 0 {
		privateNetwork, httpResp, err := c.V2Client.GetPrivateLinkNetworkConfig(consumerSharedResources[0].GetId())
		if err != nil {
			return errors.CatchCCloudV2Error(err, httpResp)
		}
		network = privateNetwork
	}

	consumerShareObj := c.buildConsumerShare(consumerShare)
	cloud := network.GetCloud()
	if cloud.CdxV1AwsNetwork == nil && cloud.CdxV1AzureNetwork == nil && cloud.CdxV1GcpNetwork == nil {
		return output.DescribeObject(cmd, consumerShareObj, consumerShareListFields, consumerHumanLabelMap, consumerStructuredLabelMap)
	}

	return c.handlePrivateLinkClusterConsumerShare(cmd, network, consumerShareObj)
}

func (c *command) handlePrivateLinkClusterConsumerShare(cmd *cobra.Command, network cdxv1.CdxV1Network, consumerShareObj *consumerShare) error {
	networkKind, privateLinkDataType, privateLinkData := getPrivateLinkNetworkDetails(network)

	consumerShareObj.NetworkDnsDomain = network.GetDnsDomain()
	consumerShareObj.NetworkZones = strings.Join(network.GetZones(), ",")
	consumerShareObj.NetworkZonalSubdomains = mapSubdomainsToList(network.GetZonalSubdomains())
	consumerShareObj.NetworkKind = networkKind
	consumerShareObj.NetworkPrivateLinkDataType = privateLinkDataType
	consumerShareObj.NetworkPrivateLinkData = privateLinkData

	return output.DescribeObject(cmd, consumerShareObj, append(consumerShareListFields, redeemTokenPrivateLinkFields...),
		combineMaps(consumerHumanLabelMap, redeemTokenPrivateLinkHumanLabelMap),
		combineMaps(consumerStructuredLabelMap, redeemTokenPrivateLinkStructuredLabelMap))
}
