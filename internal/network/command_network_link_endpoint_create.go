package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newNetworkLinkEndpointCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a network link endpoint.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.networkLinkEndpointCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a network link endpoint for network "n-123456" and network link service "nls-abcde1".`,
				Code: `confluent network network-link endpoint create --network n-123456 --description "example network link endpoint" --network-link-service nls-abcde1`,
			},
			examples.Example{
				Text: `Create a named network link endpoint for network "n-123456" and network link service "nls-abcde1".`,
				Code: `confluent network network-link endpoint create my-network-link-endpoint --network n-123456 --description "example network link endpoint" --network-link-service nls-abcde1`,
			},
		),
	}

	addNetworkFlag(cmd, c.AuthenticatedCLICommand)
	c.addNetworkLinkServiceFlag(cmd)
	cmd.Flags().String("description", "", "Network link endpoint description.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("network"))
	cobra.CheckErr(cmd.MarkFlagRequired("network-link-service"))

	return cmd
}

func (c *command) networkLinkEndpointCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	network, err := cmd.Flags().GetString("network")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	networkLinkService, err := cmd.Flags().GetString("network-link-service")
	if err != nil {
		return err
	}

	createNetworkLinkEndpoint := networkingv1.NetworkingV1NetworkLinkEndpoint{
		Spec: &networkingv1.NetworkingV1NetworkLinkEndpointSpec{
			Description:        networkingv1.PtrString(description),
			Environment:        &networkingv1.GlobalObjectReference{Id: environmentId},
			Network:            &networkingv1.EnvScopedObjectReference{Id: network},
			NetworkLinkService: &networkingv1.EnvScopedObjectReference{Id: networkLinkService},
		},
	}

	if name != "" {
		createNetworkLinkEndpoint.Spec.SetDisplayName(name)
	}

	endpoint, err := c.V2Client.CreateNetworkLinkEndpoint(createNetworkLinkEndpoint)
	if err != nil {
		return err
	}

	return printNetworkLinkEndpointTable(cmd, endpoint)
}
