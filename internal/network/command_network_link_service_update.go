package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
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
				Code: `confluent network network-link service update nls-123456 --name my-network-link-service --description "example network link service"`,
			},
			examples.Example{
				Text: `Update the accepted environments and accepted networks of network link service "nls-123456".`,
				Code: `confluent network network-link service update nls-123456 --description "example network link service" --accepted-environments env-111111 --accepted-networks n-111111,n-222222`,
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

	cmd.MarkFlagsOneRequired("name", "description", "accepted-environments", "accepted-networks")

	return cmd
}

func (c *command) networkLinkServiceUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	networkLinkService, err := c.V2Client.GetNetworkLinkService(environmentId, args[0])
	if err != nil {
		return err
	}

	updateNetworkLinkService := networkingv1.NetworkingV1NetworkLinkServiceUpdate{
		Spec: &networkingv1.NetworkingV1NetworkLinkServiceSpecUpdate{
			Environment: &networkingv1.GlobalObjectReference{Id: environmentId},
			Accept:      networkLinkService.Spec.Accept,
		},
	}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		updateNetworkLinkService.Spec.SetDisplayName(name)
	}

	if cmd.Flags().Changed("description") {
		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		updateNetworkLinkService.Spec.SetDescription(description)
	}

	if cmd.Flags().Changed("accepted-networks") {
		acceptedNetworks, err := cmd.Flags().GetStringSlice("accepted-networks")
		if err != nil {
			return err
		}
		updateNetworkLinkService.Spec.Accept.SetNetworks(acceptedNetworks)
	}

	if cmd.Flags().Changed("accepted-environments") {
		acceptedEnvironments, err := cmd.Flags().GetStringSlice("accepted-environments")
		if err != nil {
			return err
		}
		updateNetworkLinkService.Spec.Accept.SetEnvironments(acceptedEnvironments)
	}

	service, err := c.V2Client.UpdateNetworkLinkService(args[0], updateNetworkLinkService)
	if err != nil {
		return err
	}

	return printNetworkLinkServiceTable(cmd, service)
}
