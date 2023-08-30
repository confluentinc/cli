package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type listPeeringOut struct {
	Id             string `human:"ID" serialized:"id"`
	Name           string `human:"Name" serialized:"name"`
	NetworkId      string `human:"Network ID" serialized:"network_id"`
	Cloud          string `human:"Cloud" serialized:"cloud"`
	CustomRegion   string `human:"Custom Region,omitempty" serialized:"custom_region,omitempty"`
	VirtualNetwork string `human:"Virtual Nework,omitempty" serialized:"virtual_network,omitempty"`
	CloudAccount   string `human:"Cloud Account,omitempty" serialized:"cloud_account,omitempty"`
	Phase          string `human:"Phase" serialized:"phase"`
}

func (c *peeringCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List peering connections.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *peeringCommand) list(cmd *cobra.Command, _ []string) error {
	peerings, err := c.getPeerings()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, peering := range peerings {
		if peering.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if peering.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		cloud, err := getCloud(peering)
		if err != nil {
			return err
		}

		out := &listPeeringOut{
			Id:        peering.GetId(),
			Name:      peering.Spec.GetDisplayName(),
			NetworkId: peering.Spec.Network.GetId(),
			Cloud:     cloud,
			Phase:     peering.Status.GetPhase(),
		}
		switch cloud {
		case CloudAws:
			out.CustomRegion = peering.Spec.Cloud.NetworkingV1AwsPeering.GetCustomerRegion()
			out.VirtualNetwork = peering.Spec.Cloud.NetworkingV1AwsPeering.GetVpc()
			out.CloudAccount = peering.Spec.Cloud.NetworkingV1AwsPeering.GetAccount()
		case CloudGcp:
			out.VirtualNetwork = peering.Spec.Cloud.NetworkingV1GcpPeering.GetVpcNetwork()
			out.CloudAccount = peering.Spec.Cloud.NetworkingV1GcpPeering.GetProject()
		case CloudAzure:
			out.CustomRegion = peering.Spec.Cloud.NetworkingV1AzurePeering.GetCustomerRegion()
			out.VirtualNetwork = peering.Spec.Cloud.NetworkingV1AzurePeering.GetVnet()
			out.CloudAccount = peering.Spec.Cloud.NetworkingV1AzurePeering.GetTenant()
		}

		list.Add(out)
	}
	return list.Print()
}
