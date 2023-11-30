package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type networkLinkEndpointOut struct {
	Id                 string `human:"ID" serialized:"id"`
	Name               string `human:"Name,omitempty" serialized:"name,omitempty"`
	Network            string `human:"Network" serialized:"network"`
	Environment        string `human:"Environment" serialized:"environment"`
	Description        string `human:"Description,omitempty" serialized:"description,omitempty"`
	NetworkLinkService string `human:"Network Link Service" serialized:"network_link_service"`
	Phase              string `human:"Phase" serialized:"phase"`
}

func (c *command) newNetworkLinkEndpointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoint",
		Short: "Manage network link endpoints.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newNetworkLinkEndpointDescribeCommand())
	cmd.AddCommand(c.newNetworkLinkEndpointListCommand())

	return cmd
}

func (c *command) getNetworkLinkEndpoints() ([]networkingv1.NetworkingV1NetworkLinkEndpoint, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListNetworkLinkEndpoints(environmentId)
}

func (c *command) validNetworkLinkEndpointArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validNetworkLinkEndpointsArgsMultiple(cmd, args)
}

func (c *command) validNetworkLinkEndpointsArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteNetworkLinkEndpoints()
}

func (c *command) autocompleteNetworkLinkEndpoints() []string {
	endpoints, err := c.getNetworkLinkEndpoints()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(endpoints))
	for i, endpoint := range endpoints {
		suggestions[i] = fmt.Sprintf("%s\t%s", endpoint.GetId(), endpoint.Spec.GetDisplayName())
	}
	return suggestions
}

func printNetworkLinkEndpointTable(cmd *cobra.Command, endpoint networkingv1.NetworkingV1NetworkLinkEndpoint) error {
	if endpoint.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if endpoint.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	table := output.NewTable(cmd)
	table.Add(&networkLinkEndpointOut{
		Id:                 endpoint.GetId(),
		Name:               endpoint.Spec.GetDisplayName(),
		Network:            endpoint.Spec.Network.GetId(),
		Environment:        endpoint.Spec.Environment.GetId(),
		NetworkLinkService: endpoint.Spec.NetworkLinkService.GetId(),
		Description:        endpoint.Spec.GetDescription(),
		Phase:              endpoint.Status.GetPhase(),
	})

	return table.PrintWithAutoWrap(false)
}
