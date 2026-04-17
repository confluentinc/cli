package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	pcloud "github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *accessPointCommand) newIngressEndpointCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create an ingress endpoint.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.createIngressEndpoint,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an AWS PrivateLink ingress endpoint.",
				Code: "confluent network access-point private-link ingress-endpoint create --cloud aws --gateway gw-123456 --vpc-endpoint-id vpce-00000000000000000",
			},
			examples.Example{
				Text: "Create an Azure Private Link ingress endpoint.",
				Code: "confluent network access-point private-link ingress-endpoint create --cloud azure --gateway gw-123456 --private-endpoint-resource-id /subscriptions/0000000/resourceGroups/resourceGroupName/providers/Microsoft.Network/privateEndpoints/privateEndpointName",
			},
			examples.Example{
				Text: "Create a GCP Private Service Connect ingress endpoint.",
				Code: "confluent network access-point private-link ingress-endpoint create --cloud gcp --gateway gw-123456 --private-service-connect-connection-id 111111111111111111",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("vpc-endpoint-id", "", "ID of an AWS VPC endpoint.")
	cmd.Flags().String("private-endpoint-resource-id", "", "Resource ID of an Azure Private Endpoint.")
	cmd.Flags().String("private-service-connect-connection-id", "", "ID of a GCP Private Service Connect connection.")
	addGatewayFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("gateway"))

	return cmd
}

func (c *accessPointCommand) createIngressEndpoint(cmd *cobra.Command, args []string) error {
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

	vpcEndpointId, err := cmd.Flags().GetString("vpc-endpoint-id")
	if err != nil {
		return err
	}

	privateEndpointResourceId, err := cmd.Flags().GetString("private-endpoint-resource-id")
	if err != nil {
		return err
	}

	privateServiceConnectConnectionId, err := cmd.Flags().GetString("private-service-connect-connection-id")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	createIngressEndpoint := networkingaccesspointv1.NetworkingV1AccessPoint{
		Spec: &networkingaccesspointv1.NetworkingV1AccessPointSpec{
			Environment: &networkingaccesspointv1.ObjectReference{Id: environmentId},
			Gateway:     &networkingaccesspointv1.ObjectReference{Id: gateway},
		},
	}

	if name != "" {
		createIngressEndpoint.Spec.SetDisplayName(name)
	}

	switch cloud {
	case pcloud.Aws:
		if vpcEndpointId == "" {
			return fmt.Errorf("flag \"vpc-endpoint-id\" is required for AWS ingress endpoints")
		}
		createIngressEndpoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
			NetworkingV1AwsIngressPrivateLinkEndpoint: &networkingaccesspointv1.NetworkingV1AwsIngressPrivateLinkEndpoint{
				Kind:          "AwsIngressPrivateLinkEndpoint",
				VpcEndpointId: vpcEndpointId,
			},
		}
	case pcloud.Azure:
		if privateEndpointResourceId == "" {
			return fmt.Errorf("flag \"private-endpoint-resource-id\" is required for Azure ingress endpoints")
		}
		createIngressEndpoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
			NetworkingV1AzureIngressPrivateLinkEndpoint: &networkingaccesspointv1.NetworkingV1AzureIngressPrivateLinkEndpoint{
				Kind:                      "AzureIngressPrivateLinkEndpoint",
				PrivateEndpointResourceId: privateEndpointResourceId,
			},
		}
	case pcloud.Gcp:
		if privateServiceConnectConnectionId == "" {
			return fmt.Errorf("flag \"private-service-connect-connection-id\" is required for GCP ingress endpoints")
		}
		createIngressEndpoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
			NetworkingV1GcpIngressPrivateServiceConnectEndpoint: &networkingaccesspointv1.NetworkingV1GcpIngressPrivateServiceConnectEndpoint{
				Kind:                              "GcpIngressPrivateServiceConnectEndpoint",
				PrivateServiceConnectConnectionId: privateServiceConnectConnectionId,
			},
		}
	default:
		return fmt.Errorf("ingress endpoints are only supported for AWS, Azure, and GCP")
	}

	ingressEndpoint, err := c.V2Client.CreateNetworkAccessPoint(createIngressEndpoint)
	if err != nil {
		return err
	}

	return printPrivateLinkIngressEndpointTable(cmd, ingressEndpoint)
}
