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
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("vpc-endpoint-id", "", "ID of an AWS VPC endpoint.")
	addGatewayFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("gateway"))
	cobra.CheckErr(cmd.MarkFlagRequired("vpc-endpoint-id"))

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
		createIngressEndpoint.Spec.Config = &networkingaccesspointv1.NetworkingV1AccessPointSpecConfigOneOf{
			NetworkingV1AwsIngressPrivateLinkEndpoint: &networkingaccesspointv1.NetworkingV1AwsIngressPrivateLinkEndpoint{
				Kind:          "AwsIngressPrivateLinkEndpoint",
				VpcEndpointId: vpcEndpointId,
			},
		}
	default:
		return fmt.Errorf("ingress endpoints are only supported for AWS")
	}

	ingressEndpoint, err := c.V2Client.CreateAccessPoint(createIngressEndpoint)
	if err != nil {
		return err
	}

	return printPrivateLinkIngressEndpointTable(cmd, ingressEndpoint)
}
