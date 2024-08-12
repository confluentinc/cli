package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type networkLinkServiceOut struct {
	Id                   string   `human:"ID" serialized:"id"`
	Name                 string   `human:"Name,omitempty" serialized:"name,omitempty"`
	Network              string   `human:"Network" serialized:"network"`
	Environment          string   `human:"Environment" serialized:"environment"`
	Description          string   `human:"Description,omitempty" serialized:"description,omitempty"`
	AcceptedEnvironments []string `human:"Accepted Environments,omitempty" serialized:"accepted_environments,omitempty"`
	AcceptedNetworks     []string `human:"Accepted Networks,omitempty" serialized:"accepted_networks,omitempty"`
	Phase                string   `human:"Phase" serialized:"phase"`
}

func (c *command) newNetworkLinkServiceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "Manage network link services.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newNetworkLinkServiceAssociationCommand())
	cmd.AddCommand(c.newNetworkLinkServiceCreateCommand())
	cmd.AddCommand(c.newNetworkLinkServiceDeleteCommand())
	cmd.AddCommand(c.newNetworkLinkServiceDescribeCommand())
	cmd.AddCommand(c.newNetworkLinkServiceListCommand())
	cmd.AddCommand(c.newNetworkLinkServiceUpdateCommand())

	return cmd
}

func (c *command) getNetworkLinkServices(name, network, phase []string) ([]networkingv1.NetworkingV1NetworkLinkService, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListNetworkLinkServices(environmentId, name, network, phase)
}

func (c *command) validNetworkLinkServiceArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validNetworkLinkServicesArgsMultiple(cmd, args)
}

func (c *command) validNetworkLinkServicesArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteNetworkLinkServices()
}

func (c *command) autocompleteNetworkLinkServices() []string {
	services, err := c.getNetworkLinkServices(nil, nil, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(services))
	for i, service := range services {
		suggestions[i] = fmt.Sprintf("%s\t%s", service.GetId(), service.Spec.GetDisplayName())
	}
	return suggestions
}

func printNetworkLinkServiceTable(cmd *cobra.Command, service networkingv1.NetworkingV1NetworkLinkService) error {
	if service.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if service.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	table := output.NewTable(cmd)

	table.Add(&networkLinkServiceOut{
		Id:                   service.GetId(),
		Name:                 service.Spec.GetDisplayName(),
		Network:              service.Spec.Network.GetId(),
		Environment:          service.Spec.Environment.GetId(),
		Description:          service.Spec.GetDescription(),
		AcceptedEnvironments: service.Spec.Accept.GetEnvironments(),
		AcceptedNetworks:     service.Spec.Accept.GetNetworks(),
		Phase:                service.Status.GetPhase(),
	})
	return table.PrintWithAutoWrap(false)
}
