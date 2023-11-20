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
		Args:  cobra.ExactArgs(1),
		RunE:  c.networkLinkServiceCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a network link service.",
				Code: "confluent network network-link service create test_nls --network n-123456 --description 'test description' --accept-environments env-00000",
			},
		),
	}

	addNetworkFlag(cmd, c.AuthenticatedCLICommand)
	addAcceptNetworksFlag(cmd, c.AuthenticatedCLICommand)
	addAcceptEnvironmentsFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("description", "", "Network link service description.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("network"))

	return cmd
}

func (c *command) networkLinkServiceCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

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

	acceptedNetworks, err := cmd.Flags().GetStringSlice("accept-networks")
	if err != nil {
		return err
	}

	acceptedEnvironments, err := cmd.Flags().GetStringSlice("accept-environments")
	if err != nil {
		return err
	}

	createNetworkLinkService := networkingv1.NetworkingV1NetworkLinkService{
		Spec: &networkingv1.NetworkingV1NetworkLinkServiceSpec{
			DisplayName: networkingv1.PtrString(name),
			Description: networkingv1.PtrString(description),
			Environment: &networkingv1.GlobalObjectReference{Id: environmentId},
			Network:     &networkingv1.EnvScopedObjectReference{Id: networkId},
			Accept:      &networkingv1.NetworkingV1NetworkLinkServiceAcceptPolicy{},
		},
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
