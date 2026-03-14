package network

import (
	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *accessPointCommand) newPrivateNetworkInterfaceUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing private network interface.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPrivateNetworkInterfaceArgs),
		RunE:              c.privateNetworkInterfaceUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of private network interface "ap-123456".`,
				Code: "confluent network access-point private-network-interface update ap-123456 --name my-new-private-network-interface",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the private network interface.")
	cmd.Flags().StringSlice("network-interfaces", nil, "A comma-separated list of the IDs of the Elastic Network Interfaces.")
	cmd.Flags().StringSlice("routes", nil, `A comma-separated list of egress CIDR routes (max 10), e.g., "172.31.0.0/16,10.108.16.0/21".`)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("name", "network-interfaces", "routes")

	return cmd
}

func (c *accessPointCommand) privateNetworkInterfaceUpdate(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	updatePrivateNetworkInterface := networkingaccesspointv1.NetworkingV1AccessPointUpdate{
		Spec: &networkingaccesspointv1.NetworkingV1AccessPointSpecUpdate{
			DisplayName: networkingaccesspointv1.PtrString(name),
			Environment: &networkingaccesspointv1.ObjectReference{Id: environmentId},
		},
	}

	networkInterfaces, err := cmd.Flags().GetStringSlice("network-interfaces")
	if err != nil {
		return err
	}

	routes, err := cmd.Flags().GetStringSlice("routes")
	if err != nil {
		return err
	}

	if len(networkInterfaces) > 0 || len(routes) > 0 {
		awsConfig := &networkingaccesspointv1.NetworkingV1AwsPrivateNetworkInterface{
			Kind: "AwsPrivateNetworkInterface",
		}
		if len(networkInterfaces) > 0 {
			awsConfig.NetworkInterfaces = &networkInterfaces
		}
		if len(routes) > 0 {
			awsConfig.EgressRoutes = &routes
		}
		updatePrivateNetworkInterface.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecUpdateConfigOneOf{
			NetworkingV1AwsPrivateNetworkInterface: awsConfig,
		}
	}

	accessPoint, err := c.V2Client.UpdateAccessPoint(args[0], updatePrivateNetworkInterface)
	if err != nil {
		return err
	}

	return printPrivateNetworkInterfaceTable(cmd, accessPoint)
}
