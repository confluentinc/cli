package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
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
				Code: "confluent network access-point private-link egress-endpoint create --cloud aws --gateway gw-123456 --service vpce-123456 --high-availability",
			},
			examples.Example{
				Text: "Create a named Azure private link egress endpoint.",
				Code: "confluent network access-point private-link egress-endpoint create my-egress-endpoint --cloud azure --gateway gw-123456 --service vpce-123456",
			},
		),
	}

	cmd.Flags().String("cloud", "", "Specify the cloud provider as aws or azure.")
	cmd.Flags().String("service", "", "Name of an AWS VPC endpoint service or ID of an Azure Private Link service.")
	cmd.Flags().String("gateway", "", "Gateway ID.")
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

	endpoint, err := cmd.Flags().GetString("endpoint")
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

	createAccessPoint := networkingaccesspointv1.NetworkingV1AccessPoint{
		Spec: &networkingaccesspointv1.NetworkingV1AccessPointSpec{
			Environment: &networkingaccesspointv1.ObjectReference{Id: environmentId},
			Gateway:     &networkingaccesspointv1.ObjectReference{Id: gateway},
		},
	}

	if name != "" {
		createAccessPoint.Spec.SetDisplayName(name)
	}

	switch cloud {
	case CloudAws:
		createAccessPoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
			NetworkingV1AwsEgressPrivateLinkEndpoint: &networkingaccesspointv1.NetworkingV1AwsEgressPrivateLinkEndpoint{
				Kind:                   "AwsEgressPrivateLinkEndpoint",
				VpcEndpointServiceName: endpoint,
				EnableHighAvailability: networkingaccesspointv1.PtrBool(highAvailability),
			},
		}
	case CloudAzure:
		createAccessPoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
			NetworkingV1AzureEgressPrivateLinkEndpoint: &networkingaccesspointv1.NetworkingV1AzureEgressPrivateLinkEndpoint{
				Kind:                         "AzureEgressPrivateLinkEndpoint",
				PrivateLinkServiceResourceId: endpoint,
			},
		}
	}

	accessPoint, err := c.V2Client.CreateAccessPoint(createAccessPoint)
	if err != nil {
		return err
	}

	return printPrivateLinkEgressEndpointTable(cmd, accessPoint)
}
