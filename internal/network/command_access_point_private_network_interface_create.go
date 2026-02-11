package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	pcloud "github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *accessPointCommand) newPrivateNetworkInterfaceCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a private network interface.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.privateNetworkInterfaceCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an AWS private network interface access point.",
				Code: "confluent network access-point private-network-interface create --cloud aws --gateway gw-123456 --network-interfaces usw2-az1,usw2-az2,usw2-az3 --account 000000000000",
			},
		),
	}

	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString([]string{"aws"}, "or")))
	addGatewayFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().StringSlice("network-interfaces", nil, "A comma-separated list of the IDs of the Elastic Network Interfaces.")
	cmd.Flags().String("account", "", "The AWS account ID associated with the Elastic Network Interfaces.")
	cmd.Flags().StringSlice("routes", nil, `A comma-separated list of egress CIDR routes (max 10), e.g., "172.31.0.0/16,10.108.16.0/21".`)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("gateway"))
	cobra.CheckErr(cmd.MarkFlagRequired("network-interfaces"))
	cobra.CheckErr(cmd.MarkFlagRequired("account"))

	return cmd
}

func (c *accessPointCommand) privateNetworkInterfaceCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)

	gateway, err := cmd.Flags().GetString("gateway")
	if err != nil {
		return err
	}

	account, err := cmd.Flags().GetString("account")
	if err != nil {
		return err
	}

	networkInterfaces, err := cmd.Flags().GetStringSlice("network-interfaces")
	if err != nil {
		return err
	}

	routes, err := cmd.Flags().GetStringSlice("routes")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	createPrivateNetworkInterface := networkingaccesspointv1.NetworkingV1AccessPoint{
		Spec: &networkingaccesspointv1.NetworkingV1AccessPointSpec{
			Environment: &networkingaccesspointv1.ObjectReference{Id: environmentId},
			Gateway:     &networkingaccesspointv1.ObjectReference{Id: gateway},
		},
	}

	if name != "" {
		createPrivateNetworkInterface.Spec.SetDisplayName(name)
	}

	switch cloud {
	case pcloud.Aws:
		awsConfig := &networkingaccesspointv1.NetworkingV1AwsPrivateNetworkInterface{
			Kind:              "AwsPrivateNetworkInterface",
			NetworkInterfaces: &networkInterfaces,
			Account:           &account,
		}
		if len(routes) > 0 {
			awsConfig.EgressRoutes = &routes
		}
		createPrivateNetworkInterface.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
			NetworkingV1AwsPrivateNetworkInterface: awsConfig,
		}
	}

	privateNetworkInterface, err := c.V2Client.CreateAccessPoint(createPrivateNetworkInterface)
	if err != nil {
		return err
	}

	return printPrivateNetworkInterfaceTable(cmd, privateNetworkInterface)
}
