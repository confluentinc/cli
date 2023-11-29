package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newNetworkLinkServiceUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing network link service.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validNetworkLinkServiceArgs),
		RunE:              c.networkLinkServiceUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name and description of network link service "nls-123456".`,
				Code: `confluent network network-link service update nls-123456 --name my-network-link-service --description "example network link service"'`,
			},
			examples.Example{
				Text: `Update the accepted environments and accepted networks of network link service "nls-123456".`,
				Code: `confluent network network-link service update nls-123456 --description "example network link service" --accepted-environments env-111111 --accepted-networks n-111111,n222222`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the network link service.")
	cmd.Flags().String("description", "", "Description of the network link service.")
	addAcceptedNetworksFlag(cmd, c.AuthenticatedCLICommand)
	addAcceptedEnvironmentsFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) networkLinkServiceUpdate(cmd *cobra.Command, args []string) error {
	if err := errors.CheckNoUpdate(cmd.Flags(), "name", "description", "accepted-environments", "accepted-networks"); err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	updateNetworkLinkService := networkingv1.NetworkingV1NetworkLinkServiceUpdate{
		Spec: &networkingv1.NetworkingV1NetworkLinkServiceSpecUpdate{
			DisplayName: networkingv1.PtrString(name),
			Description: networkingv1.PtrString(description),
			Environment: &networkingv1.GlobalObjectReference{Id: environmentId},
		},
	}

	if cmd.Flags().Changed("accepted-networks") || cmd.Flags().Changed("accepted-environments") {
		updateNetworkLinkService.Spec.Accept = &networkingv1.NetworkingV1NetworkLinkServiceAcceptPolicy{}

		acceptedNetworks, err := cmd.Flags().GetStringSlice("accepted-networks")
		if err != nil {
			return err
		}

		acceptedEnvironments, err := cmd.Flags().GetStringSlice("accepted-environments")
		if err != nil {
			return err
		}

		networkLinkService, err := c.V2Client.GetNetworkLinkService(environmentId, args[0])
		if err != nil {
			return err
		}

		if cmd.Flags().Changed("accepted-networks") && cmd.Flags().Changed("accepted-environments") {
			updateNetworkLinkService.Spec.Accept.SetNetworks(acceptedNetworks)
			updateNetworkLinkService.Spec.Accept.SetEnvironments(acceptedEnvironments)
		} else if cmd.Flags().Changed("accepted-networks") {
			updateNetworkLinkService.Spec.Accept.SetNetworks(acceptedNetworks)
			updateNetworkLinkService.Spec.Accept.Environments = networkLinkService.Spec.Accept.Environments
		} else {
			updateNetworkLinkService.Spec.Accept.Networks = networkLinkService.Spec.Accept.Networks
			updateNetworkLinkService.Spec.Accept.SetEnvironments(acceptedEnvironments)
		}
	}

	service, err := c.V2Client.UpdateNetworkLinkService(args[0], updateNetworkLinkService)
	if err != nil {
		return err
	}

	return printNetworkLinkServiceTable(cmd, service)
}
