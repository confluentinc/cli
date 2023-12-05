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

	list := output.NewList(cmd)
	for _, endpoint := range endpoints {
		if endpoint.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if endpoint.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
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
