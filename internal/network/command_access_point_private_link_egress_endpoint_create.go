package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func (c *accessPointCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create an egress endpoint.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an AWS private link egress endpoint with high availability.",
				Code: "confluent network access-point private-link egress-endpoint create --cloud aws --gateway gw-123456 --service com.amazonaws.vpce.us-west-2.vpce-svc-00000000000000000 --high-availability",
			},
			examples.Example{
				Text: `Create an Azure Private Link egress endpoint named "my-egress-endpoint".`,
				Code: "confluent network access-point private-link egress-endpoint create my-egress-endpoint --cloud azure --gateway gw-123456 --service /subscriptions/0000000/resourceGroups/plsRgName/providers/Microsoft.Network/privateLinkServices/privateLinkServiceName",
			},
		),
	}

	cmd.Flags().String("cloud", "", fmt.Sprintf("Specify the cloud provider as %s.", utils.ArrayToCommaDelimitedString([]string{"aws", "azure"}, "or")))
	cmd.Flags().String("service", "", "Name of an AWS VPC endpoint service or ID of an Azure Private Link service.")
	addGatewayFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("subresource", "", "Name of an Azure Private Link subresource.")
	cmd.Flags().Bool("high-availability", false, "Enable high availability for AWS egress endpoint.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("gateway"))
	cobra.CheckErr(cmd.MarkFlagRequired("service"))

	return cmd
}

func (c *accessPointCommand) create(cmd *cobra.Command, args []string) error {
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

	service, err := cmd.Flags().GetString("service")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	highAvailability, err := cmd.Flags().GetBool("high-availability")
	if err != nil {
		return err
	}

	subresource, err := cmd.Flags().GetString("subresource")
	if err != nil {
		return err
	}

	createEgressEndpoint := networkingaccesspointv1.NetworkingV1AccessPoint{
		Spec: &networkingaccesspointv1.NetworkingV1AccessPointSpec{
			Environment: &networkingaccesspointv1.ObjectReference{Id: environmentId},
			Gateway:     &networkingaccesspointv1.ObjectReference{Id: gateway},
		},
	}

	if name != "" {
		createEgressEndpoint.Spec.SetDisplayName(name)
	}

	switch cloud {
	case CloudAws:
		createEgressEndpoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
			NetworkingV1AwsEgressPrivateLinkEndpoint: &networkingaccesspointv1.NetworkingV1AwsEgressPrivateLinkEndpoint{
				Kind:                   "AwsEgressPrivateLinkEndpoint",
				VpcEndpointServiceName: service,
				EnableHighAvailability: networkingaccesspointv1.PtrBool(highAvailability),
			},
		}
	case CloudAzure:
		createEgressEndpoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
			NetworkingV1AzureEgressPrivateLinkEndpoint: &networkingaccesspointv1.NetworkingV1AzureEgressPrivateLinkEndpoint{
				Kind:                         "AzureEgressPrivateLinkEndpoint",
				PrivateLinkServiceResourceId: service,
			},
		}
		if subresource != "" {
			createEgressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint.PrivateLinkSubresourceName = networkingaccesspointv1.PtrString(subresource)
		}
	}

	egressEndpoint, err := c.V2Client.CreateAccessPoint(createEgressEndpoint)
	if err != nil {
		return err
	}

	return printPrivateLinkEgressEndpointTable(cmd, egressEndpoint)
}
