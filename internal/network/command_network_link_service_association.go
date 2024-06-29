package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type networkLinkServiceAssociationOut struct {
	Id                  string `human:"ID" serialized:"id"`
	Name                string `human:"Name" serialized:"name"`
	Environment         string `human:"Environment" serialized:"environment"`
	Description         string `human:"Description,omitempty" serialized:"description,omitempty"`
	NetworkLinkEndpoint string `human:"Network Link Endpoint" serialized:"network_link_endpoint"`
	NetworkLinkService  string `human:"Network Link Service" serialized:"network_link_service"`
	Phase               string `human:"Phase" serialized:"phase"`
}

func (c *command) newNetworkLinkServiceAssociationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "association",
		Short: "Manage network link service associations.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newNetworkLinkServiceAssociationDescribeCommand())
	cmd.AddCommand(c.newNetworkLinkServiceAssociationListCommand())

	return cmd
}

func printNetworkLinkServiceAssociationTable(cmd *cobra.Command, association networkingv1.NetworkingV1NetworkLinkServiceAssociation) error {
	if association.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if association.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	table := output.NewTable(cmd)
	table.Add(&networkLinkServiceAssociationOut{
		Id:                  association.GetId(),
		Name:                association.Spec.GetDisplayName(),
		Environment:         association.Spec.Environment.GetId(),
		Description:         association.Spec.GetDescription(),
		NetworkLinkEndpoint: association.Spec.GetNetworkLinkEndpoint(),
		NetworkLinkService:  association.Spec.NetworkLinkService.GetId(),
		Phase:               association.Status.GetPhase(),
	})

	return table.Print()
}
