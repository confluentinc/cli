package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newNetworkLinkServiceAssociationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List associations for a network link service.",
		Args:  cobra.NoArgs,
		RunE:  c.networkLinkServiceAssociationList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List associations for network link service "nls-123456".`,
				Code: "confluent network network-link service association list --network-link-service nls-123456",
			},
		),
	}

	c.addNetworkLinkServiceFlag(cmd)
	addPhaseFlag(cmd, resource.NetworkLinkServiceAssociation)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("network-link-service"))

	return cmd
}

func (c *command) networkLinkServiceAssociationList(cmd *cobra.Command, _ []string) error {
	networkLinkService, err := cmd.Flags().GetString("network-link-service")
	if err != nil {
		return err
	}

	phase, err := cmd.Flags().GetStringSlice("phase")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	associations, err := c.V2Client.ListNetworkLinkServiceAssociations(environmentId, networkLinkService, phase)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, association := range associations {
		if association.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if association.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		list.Add(&networkLinkServiceAssociationOut{
			Id:                  association.GetId(),
			Name:                association.Spec.GetDisplayName(),
			Description:         association.Spec.GetDescription(),
			NetworkLinkEndpoint: association.Spec.GetNetworkLinkEndpoint(),
			NetworkLinkService:  association.Spec.NetworkLinkService.GetId(),
			Phase:               association.Status.GetPhase(),
		})
	}

	list.Filter([]string{"Id", "Name", "Description", "NetworkLinkEndpoint", "NetworkLinkService", "Phase"})

	return list.Print()
}
