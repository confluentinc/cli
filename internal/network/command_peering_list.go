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
	Network        string `human:"Network" serialized:"network"`
	Cloud          string `human:"Cloud" serialized:"cloud"`
	CustomRegion   string `human:"Custom Region,omitempty" serialized:"custom_region,omitempty"`
	VirtualNetwork string `human:"Virtual Nework,omitempty" serialized:"virtual_network,omitempty"`
	CloudAccount   string `human:"Cloud Account,omitempty" serialized:"cloud_account,omitempty"`
	Phase          string `human:"Phase" serialized:"phase"`
}

func (c *command) newPeeringListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List peerings.",
		Args:  cobra.NoArgs,
		RunE:  c.peeringList,
	}

	cmd.Flags().StringSlice("name", nil, "A comma-separated list of peering names.")
	addListNetworkFlag(cmd, c.AuthenticatedCLICommand)
	addPeeringPhaseFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) peeringList(cmd *cobra.Command, _ []string) error {
	name, err := cmd.Flags().GetStringSlice("name")
	if err != nil {
		return err
	}

	network, err := cmd.Flags().GetStringSlice("network")
	if err != nil {
		return err
	}

	phase, err := cmd.Flags().GetStringSlice("phase")
	if err != nil {
		return err
	}

	peerings, err := c.getPeerings(name, network, phase)
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

		cloud, err := getPeeringCloud(peering)
		if err != nil {
			return err
		}

		out := &listPeeringOut{
			Id:      peering.GetId(),
			Name:    peering.Spec.GetDisplayName(),
			Network: peering.Spec.Network.GetId(),
			Cloud:   cloud,
			Phase:   peering.Status.GetPhase(),
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
