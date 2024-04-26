package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newNetworkLinkServiceCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a network link service.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.networkLinkServiceCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a network link service for network "n-123456" with accepted environments "env-111111" and "env-222222".`,
				Code: `confluent network link service create --network n-123456 --description "example network link service" --accepted-environments env-111111,env-222222`,
			},
			examples.Example{
				Text: `Create a named network link service for network "n-123456" with accepted networks "n-abced1" and "n-abcde2".`,
				Code: `confluent network link service create my-network-link-service --network n-123456 --description "example network link service" --accepted-networks n-abcde1,n-abcde2`,
			},
		),
	}

	addNetworkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("description", "", "Network link service description.")
	addAcceptedNetworksFlag(cmd, c.AuthenticatedCLICommand)
	addAcceptedEnvironmentsFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("network"))
	cmd.MarkFlagsOneRequired("accepted-networks", "accepted-environments")

	return cmd
}

func (c *command) networkLinkServiceCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	networkId, err := cmd.Flags().GetString("network")
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

	acceptedNetworks, err := cmd.Flags().GetStringSlice("accepted-networks")
	if err != nil {
		return err
	}

	acceptedEnvironments, err := cmd.Flags().GetStringSlice("accepted-environments")
	if err != nil {
		return err
	}

	createNetworkLinkService := networkingv1.NetworkingV1NetworkLinkService{
		Spec: &networkingv1.NetworkingV1NetworkLinkServiceSpec{
			Description: networkingv1.PtrString(description),
			Environment: &networkingv1.GlobalObjectReference{Id: environmentId},
			Network:     &networkingv1.EnvScopedObjectReference{Id: networkId},
			Accept:      &networkingv1.NetworkingV1NetworkLinkServiceAcceptPolicy{},
		},
	}

	if name != "" {
		createNetworkLinkService.Spec.SetDisplayName(name)
	}

	if len(acceptedNetworks) > 0 {
		createNetworkLinkService.Spec.Accept.SetNetworks(acceptedNetworks)
	}

	if len(acceptedEnvironments) > 0 {
		createNetworkLinkService.Spec.Accept.SetEnvironments(acceptedEnvironments)
	}

	service, err := c.V2Client.CreateNetworkLinkService(createNetworkLinkService)
	if err != nil {
		return err
	}

	return printNetworkLinkServiceTable(cmd, service)
}
