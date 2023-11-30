package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type listNetworkLinkEndpointOut struct {
	Id                 string `human:"ID" serialized:"id"`
	Name               string `human:"Name" serialized:"name"`
	Network            string `human:"Network" serialized:"network"`
	Description        string `human:"Description,omitempty" serialized:"description,omitempty"`
	NetworkLinkService string `human:"Network Link Service" serialized:"network_link_service"`
	Phase              string `human:"Phase" serialized:"phase"`
}

func (c *command) newNetworkLinkEndpointListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List network link endpoints.",
		Args:  cobra.NoArgs,
		RunE:  c.networkLinkEndpointList,
	}

	cmd.Flags().String("name", "", "Network Link endpoint display name.")
	addNetworkFlag(cmd, c.AuthenticatedCLICommand)
	c.addNetworkLinkServiceFlag(cmd)
	addNetworkLinkEndpointPhaseFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) networkLinkEndpointList(cmd *cobra.Command, _ []string) error {
	endpoints, err := c.getNetworkLinkEndpoints()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	network, err := cmd.Flags().GetString("network")
	if err != nil {
		return err
	}

	service, err := cmd.Flags().GetString("network-link-service")
	if err != nil {
		return err
	}

	phase, err := cmd.Flags().GetString("phase")
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, endpoint := range endpoints {
		if endpoint.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if endpoint.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		if name != "" && endpoint.Spec.GetDisplayName() != name {
			continue
		}
		if network != "" && endpoint.Spec.Network.GetId() != network {
			continue
		}
		if service != "" && endpoint.Spec.NetworkLinkService.GetId() != service {
			continue
		}
		if phase != "" && endpoint.Status.GetPhase() != phase {
			continue
		}

		list.Add(&listNetworkLinkEndpointOut{
			Id:                 endpoint.GetId(),
			Name:               endpoint.Spec.GetDisplayName(),
			Network:            endpoint.Spec.Network.GetId(),
			NetworkLinkService: endpoint.Spec.NetworkLinkService.GetId(),
			Description:        endpoint.Spec.GetDescription(),
			Phase:              endpoint.Status.GetPhase(),
		})
	}
	return list.Print()
}
