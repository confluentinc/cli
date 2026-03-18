package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkinggatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking-gateway/v1"

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
				Text: `Create AWS ingress private link gateway "my-ingress-gateway".`,
				Code: "confluent network gateway create my-ingress-gateway --cloud aws --region us-east-1 --type ingress-privatelink",
			},
			examples.Example{
				Text: `Create AWS egress private link gateway "my-egress-gateway".`,
				Code: "confluent network gateway create my-egress-gateway --cloud aws --region us-east-1 --type egress-privatelink",
			},
			examples.Example{
				Text: `Create AWS private network interface gateway "my-pni-gateway".`,
				Code: "confluent network gateway create my-pni-gateway --cloud aws --region us-east-1 --type private-network-interface",
			},
			examples.Example{
				Text: `Create Azure ingress private link gateway "my-azure-ingress-gateway".`,
				Code: "confluent network gateway create my-azure-ingress-gateway --cloud azure --region eastus2 --type ingress-privatelink",
			},
			examples.Example{
				Text: `Create GCP ingress private service connect gateway "my-gcp-ingress-gateway".`,
				Code: "confluent network gateway create my-gcp-ingress-gateway --cloud gcp --region us-central1 --type ingress-private-service-connect",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	addGatewayTypeFlag(cmd)
	c.addRegionFlagGateway(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().StringSlice("zones", nil, "A comma-separated list of availability zones for this gateway.")
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

	zones, err := cmd.Flags().GetStringSlice("zones")
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
		} else if gatewayType == "ingress-privatelink" {
			createGateway.Spec.Config = &networkinggatewayv1.NetworkingV1GatewaySpecConfigOneOf{
				NetworkingV1AwsIngressPrivateLinkGatewaySpec: &networkinggatewayv1.NetworkingV1AwsIngressPrivateLinkGatewaySpec{
					Kind:   "AwsIngressPrivateLinkGatewaySpec",
					Region: region,
				},
			}
		} else if gatewayType == "private-network-interface" {
			createGateway.Spec.Config = &networkinggatewayv1.NetworkingV1GatewaySpecConfigOneOf{
				NetworkingV1AwsPrivateNetworkInterfaceGatewaySpec: &networkinggatewayv1.NetworkingV1AwsPrivateNetworkInterfaceGatewaySpec{
					Kind:   "AwsPrivateNetworkInterfaceGatewaySpec",
					Region: region,
					Zones:  zones,
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
		} else if gatewayType == "ingress-privatelink" {
			createGateway.Spec.Config = &networkinggatewayv1.NetworkingV1GatewaySpecConfigOneOf{
				NetworkingV1AzureIngressPrivateLinkGatewaySpec: &networkinggatewayv1.NetworkingV1AzureIngressPrivateLinkGatewaySpec{
					Kind:   "AzureIngressPrivateLinkGatewaySpec",
					Region: region,
				},
			}
		}
	case pcloud.Gcp:
		if gatewayType == "ingress-private-service-connect" {
			createGateway.Spec.Config = &networkinggatewayv1.NetworkingV1GatewaySpecConfigOneOf{
				NetworkingV1GcpIngressPrivateServiceConnectGatewaySpec: &networkinggatewayv1.NetworkingV1GcpIngressPrivateServiceConnectGatewaySpec{
					Kind:   "GcpIngressPrivateServiceConnectGatewaySpec",
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
