package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

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

		list.Add(&networkLinkEndpointOut{
			Id:                 endpoint.GetId(),
			Name:               endpoint.Spec.GetDisplayName(),
			Network:            endpoint.Spec.Network.GetId(),
			Environment:        endpoint.Spec.Environment.GetId(),
			Description:        endpoint.Spec.GetDescription(),
			NetworkLinkService: endpoint.Spec.NetworkLinkService.GetId(),
			Phase:              endpoint.Status.GetPhase(),
		})
	}
	list.Filter([]string{"Id", "Name", "Network", "Description", "NetworkLinkService", "Phase"})
	return list.Print()
}
