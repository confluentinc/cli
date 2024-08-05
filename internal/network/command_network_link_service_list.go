package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newNetworkLinkServiceListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List network link services.",
		Args:  cobra.NoArgs,
		RunE:  c.networkLinkServiceList,
	}

	cmd.Flags().StringSlice("name", nil, "A comma-separated list of network link service names.")
	addListNetworkFlag(cmd, c.AuthenticatedCLICommand)
	addPhaseFlag(cmd, resource.NetworkLinkService)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) networkLinkServiceList(cmd *cobra.Command, _ []string) error {
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

	phase = toUpper(phase)

	services, err := c.getNetworkLinkServices(name, network, phase)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, service := range services {
		if service.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if service.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		list.Add(&networkLinkServiceOut{
			Id:                   service.GetId(),
			Name:                 service.Spec.GetDisplayName(),
			Network:              service.Spec.Network.GetId(),
			Description:          service.Spec.GetDescription(),
			AcceptedEnvironments: service.Spec.Accept.GetEnvironments(),
			AcceptedNetworks:     service.Spec.Accept.GetNetworks(),
			Phase:                service.Status.GetPhase(),
		})
	}
	list.Filter([]string{"Id", "Name", "Network", "Description", "AcceptedEnvironments", "AcceptedNetworks", "Phase"})
	return list.Print()
}
