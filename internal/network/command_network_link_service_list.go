package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type listNetworkLinkServiceHumanOut struct {
	Id                   string `human:"ID"`
	Name                 string `human:"Name"`
	NetworkId            string `human:"Network ID"`
	Description          string `human:"Description,omitempty"`
	AcceptedEnvironments string `human:"Accepted Environments,omitempty"`
	AcceptedNetworks     string `human:"Accepted Networks,omitempty"`
	Phase                string `human:"Phase"`
}

type listNetworkLinkServiceSerializedOut struct {
	Id                   string   `serialized:"id"`
	Name                 string   `serialized:"name"`
	NetworkId            string   `serialized:"network_id"`
	Description          string   `serialized:"description,omitempty"`
	AcceptedEnvironments []string `serialized:"accepted_environments,omitempty"`
	AcceptedNetworks     []string `serialized:"accepted_networks,omitempty"`
	Phase                string   `serialized:"phase"`
}

func (c *command) newNetworkLinkServiceListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List network link services.",
		Args:  cobra.NoArgs,
		RunE:  c.networkLinkServiceList,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) networkLinkServiceList(cmd *cobra.Command, _ []string) error {
	services, err := c.getNetworkLinkServices()
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

		if output.GetFormat(cmd) == output.Human {
			list.Add(&listNetworkLinkServiceHumanOut{
				Id:                   service.GetId(),
				Name:                 service.Spec.GetDisplayName(),
				NetworkId:            service.Spec.Network.GetId(),
				Description:          service.Spec.GetDescription(),
				AcceptedEnvironments: strings.Join(service.Spec.Accept.GetEnvironments(), ", "),
				AcceptedNetworks:     strings.Join(service.Spec.Accept.GetNetworks(), ", "),
				Phase:                service.Status.GetPhase(),
			})
		} else {
			list.Add(&listNetworkLinkServiceSerializedOut{
				Id:                   service.GetId(),
				Name:                 service.Spec.GetDisplayName(),
				NetworkId:            service.Spec.Network.GetId(),
				Description:          service.Spec.GetDescription(),
				AcceptedEnvironments: service.Spec.Accept.GetEnvironments(),
				AcceptedNetworks:     service.Spec.Accept.GetNetworks(),
				Phase:                service.Status.GetPhase(),
			})
		}
	}
	return list.Print()
}
