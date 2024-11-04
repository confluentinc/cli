package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkinggatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-gateway/v1"

	pcloud "github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newGatewayCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a network gateway.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.gatewayCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create network gateway "my-gateway".`,
				Code: "confluent network gateway create my-gateway --cloud aws --region us-east-1 --type egress-privatelink",
			},
		),
	}

	pcmd.AddCloudAwsAzureFlag(cmd)
	addGatewayTypeFlag(cmd)
	c.addRegionFlagGateway(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("type"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) gatewayCreate(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	gatewayType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	createGateway := networkinggatewayv1.NetworkingV1Gateway{
		Spec: &networkinggatewayv1.NetworkingV1GatewaySpec{
			Environment: &networkinggatewayv1.ObjectReference{Id: environmentId},
		},
	}

	switch cloud {
	case pcloud.Aws:
		if gatewayType == "egress-privatelink" {
			createGateway.Spec.Config = &networkinggatewayv1.NetworkingV1GatewaySpecConfigOneOf{
				NetworkingV1AwsEgressPrivateLinkGatewaySpec: &networkinggatewayv1.NetworkingV1AwsEgressPrivateLinkGatewaySpec{
					Kind:   "AwsEgressPrivateLinkGatewaySpec",
					Region: region,
				},
			}
		}
	case pcloud.Azure:
		if gatewayType == "egress-privatelink" {
			createGateway.Spec.Config = &networkinggatewayv1.NetworkingV1GatewaySpecConfigOneOf{
				NetworkingV1AzureEgressPrivateLinkGatewaySpec: &networkinggatewayv1.NetworkingV1AzureEgressPrivateLinkGatewaySpec{
					Kind:   "AzureEgressPrivateLinkGatewaySpec",
					Region: region,
				},
			}
		}
	case pcloud.Gcp:
		if gatewayType == "egress-privatelink" {
			createGateway.Spec.Config = &networkinggatewayv1.NetworkingV1GatewaySpecConfigOneOf{
				NetworkingV1GcpEgressPrivateServiceConnectGatewaySpec: &networkinggatewayv1.NetworkingV1GcpEgressPrivateServiceConnectGatewaySpec{
					Kind:   "GcpEgressPrivateLinkGatewaySpec",
					Region: region,
				},
			}
		}
	}

	if len(args) == 1 {
		createGateway.Spec.SetDisplayName(args[0])
	}

	gateway, err := c.V2Client.CreateGateway(createGateway)
	if err != nil {
		return err
	}

	return printGatewayTable(cmd, gateway)
}
