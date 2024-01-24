package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newNetworkLinkEndpointUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing network link endpoint.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validNetworkLinkEndpointArgs),
		RunE:              c.networkLinkEndpointUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name and description of network link endpoint "nle-123456".`,
				Code: `confluent network network-link endpoint update nle-123456 --name my-network-link-endpoint --description "example network link endpoint"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the network link endpoint.")
	cmd.Flags().String("description", "", "Description of the network link endpoint.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("name", "description")

	return cmd
}

func (c *command) networkLinkEndpointUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	updateNetworkLinkEndpoint := networkingv1.NetworkingV1NetworkLinkEndpointUpdate{
		Spec: &networkingv1.NetworkingV1NetworkLinkEndpointSpecUpdate{
			Environment: &networkingv1.GlobalObjectReference{Id: environmentId},
		},
	}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		updateNetworkLinkEndpoint.Spec.SetDisplayName(name)
	}

	if cmd.Flags().Changed("description") {
		description, err := cmd.Flags().GetString("description")
		if err != nil {
			return err
		}
		updateNetworkLinkEndpoint.Spec.SetDescription(description)
	}

	endpoint, err := c.V2Client.UpdateNetworkLinkEndpoint(args[0], updateNetworkLinkEndpoint)
	if err != nil {
		return err
	}

	return printNetworkLinkEndpointTable(cmd, endpoint)
}
