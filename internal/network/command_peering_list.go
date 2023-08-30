package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type listPeeringHumanOut struct {
	Id             string `human:"ID"`
	Name           string `human:"Name"`
	NetworkId      string `human:"Network ID"`
	Cloud          string `human:"Cloud"`
	CustomRegion   string `human:"Custom Region,omitempty"`
	VirtualNetwork string `human:"Virtual Netwrok, omitempty"`
	CloudAccount   string `human:"Cloud Account, omitempty"`
	Phase          string `human:"Phase"`
}

type listPeeringSerializedOut struct {
	Id             string `serialized:"id"`
	Name           string `serialized:"name"`
	NetworkId      string `serialized:"network_id"`
	Cloud          string `serialized:"cloud"`
	CustomRegion   string `serialized:"custom_region,omitempty"`
	VirtualNetwork string `serialized:"virtual_network, omitempty"`
	CloudAccount   string `serialized:"cloud_account, omitempty"`
	Phase          string `serialized:"phase"`
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

		human := &listPeeringHumanOut{
			Id:        peering.GetId(),
			Name:      peering.Spec.GetDisplayName(),
			NetworkId: peering.Spec.Network.GetId(),
			Cloud:     cloud,
			Phase:     peering.Status.GetPhase(),
		}

		serialized := &listPeeringSerializedOut{
			Id:        peering.GetId(),
			Name:      peering.Spec.GetDisplayName(),
			NetworkId: peering.Spec.Network.GetId(),
			Cloud:     cloud,
			Phase:     peering.Status.GetPhase(),
		}

		switch cloud {
		case CloudAws:
			human.CustomRegion = peering.Spec.Cloud.NetworkingV1AwsPeering.GetCustomerRegion()
			human.VirtualNetwork = peering.Spec.Cloud.NetworkingV1AwsPeering.GetVpc()
			human.CloudAccount = peering.Spec.Cloud.NetworkingV1AwsPeering.GetAccount()
			serialized.CustomRegion = peering.Spec.Cloud.NetworkingV1AwsPeering.GetCustomerRegion()
			serialized.VirtualNetwork = peering.Spec.Cloud.NetworkingV1AwsPeering.GetVpc()
			serialized.CloudAccount = peering.Spec.Cloud.NetworkingV1AwsPeering.GetAccount()
		case CloudGcp:
			human.VirtualNetwork = peering.Spec.Cloud.NetworkingV1GcpPeering.GetVpcNetwork()
			human.CloudAccount = peering.Spec.Cloud.NetworkingV1GcpPeering.GetProject()
			serialized.VirtualNetwork = peering.Spec.Cloud.NetworkingV1GcpPeering.GetVpcNetwork()
			serialized.CloudAccount = peering.Spec.Cloud.NetworkingV1GcpPeering.GetProject()
		case CloudAzure:
			human.CustomRegion = peering.Spec.Cloud.NetworkingV1AzurePeering.GetCustomerRegion()
			human.VirtualNetwork = peering.Spec.Cloud.NetworkingV1AzurePeering.GetVnet()
			human.CloudAccount = peering.Spec.Cloud.NetworkingV1AzurePeering.GetTenant()
			serialized.CustomRegion = peering.Spec.Cloud.NetworkingV1AzurePeering.GetCustomerRegion()
			serialized.VirtualNetwork = peering.Spec.Cloud.NetworkingV1AzurePeering.GetVnet()
			serialized.CloudAccount = peering.Spec.Cloud.NetworkingV1AzurePeering.GetTenant()
		}

		if output.GetFormat(cmd) == output.Human {
			list.Add(human)
		} else {
			list.Add(serialized)
		}
	}
	return list.Print()
}
