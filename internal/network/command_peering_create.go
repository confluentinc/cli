package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *peeringCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a peering.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an AWS VPC peering.",
				Code: "confluent network peering create aws-peering --network n-abcde1 --cloud aws --aws-account 012345678901 --aws-vpc vpc-abcdef0123456789a --aws-routes 172.31.0.0/16,10.108.16.0/21",
			},
			examples.Example{
				Text: "Create a GCP VPC peering.",
				Code: "confluent network peering create gcp-peering --network n-abcde1 --cloud gcp --gcp-project temp-123456 --gcp-vpc-network customer-test-vpc-network --gcp-import-custom-routes",
			},
			examples.Example{
				Text: "Create an Azure VNet peering.",
				Code: "confluent network peering create azure-peering --network n-abcde1 --cloud azure --azure-tenant 1111tttt-1111-1111-1111-111111tttttt --azure-vnet /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet --customer-region centralus",
			},
		),
	}

	addNetworkFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("aws-account", "", "AWS account ID associated with the VPC that you are peering with Confluent Cloud network.")
	cmd.Flags().String("gcp-project", "", "Google Cloud project ID associated with the VPC that you are peering with Confluent Cloud network.")
	cmd.Flags().String("azure-tenant", "", "Azure Tenant ID in which your Azure Subscription exists.")
	cmd.Flags().String("aws-vpc", "", "AWS VPC ID that you are peering with Confluent Cloud network.")
	cmd.Flags().String("gcp-vpc-network", "", "Name of the Google Cloud VPC that you are peering with Confluent Cloud network.")
	cmd.Flags().String("azure-vnet", "", "Azure Resource ID of the VNet that you are peering with Confluent Cloud.")
	cmd.Flags().StringSlice("aws-routes", nil, "A comma-separated list of CIDR blocks of the AWS VPC that you are peering with Confluent Cloud network.")
	cmd.Flags().Bool("gcp-import-custom-routes", false, "Enable customer route import for Google Cloud VPC Peering.")
	cmd.Flags().String("customer-region", "", "Cloud region ID of the AWS VPC or Azure VNet that you are peering with Confluent Cloud network.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("network"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))

	cmd.MarkFlagsRequiredTogether("aws-account", "aws-vpc", "aws-routes")
	cmd.MarkFlagsRequiredTogether("gcp-project", "gcp-vpc-network")
	cmd.MarkFlagsRequiredTogether("azure-tenant", "azure-vnet")

	return cmd
}

func (c *peeringCommand) create(cmd *cobra.Command, args []string) error {
	name := args[0]

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)

	networkId, err := cmd.Flags().GetString("network")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	network, err := c.V2Client.GetNetwork(environmentId, networkId)
	if err != nil {
		return err
	}
	region := network.Spec.GetRegion()

	createPeering := networkingv1.NetworkingV1Peering{
		Spec: &networkingv1.NetworkingV1PeeringSpec{
			Cloud:       &networkingv1.NetworkingV1PeeringSpecCloudOneOf{},
			DisplayName: networkingv1.PtrString(name),
			Environment: &networkingv1.ObjectReference{Id: environmentId},
			Network:     &networkingv1.ObjectReference{Id: networkId},
		},
	}

	switch cloud {
	case CloudAws:
		awsPeering, err := c.createAwsPeeringRequest(cmd, region)
		if err != nil {
			return err
		}
		createPeering.Spec.Cloud.NetworkingV1AwsPeering = awsPeering
	case CloudGcp:
		gcpPeering, err := c.createGcpPeeringRequest(cmd)
		if err != nil {
			return err
		}
		createPeering.Spec.Cloud.NetworkingV1GcpPeering = gcpPeering
	case CloudAzure:
		azurePeering, err := c.createAzurePeeringRequest(cmd, region)
		if err != nil {
			return err
		}
		createPeering.Spec.Cloud.NetworkingV1AzurePeering = azurePeering
	}

	peering, err := c.V2Client.CreatePeering(createPeering)
	if err != nil {
		return err
	}

	return printPeeringTable(cmd, peering)
}

func (c *peeringCommand) createAwsPeeringRequest(cmd *cobra.Command, networkRegion string) (*networkingv1.NetworkingV1AwsPeering, error) {
	account, err := cmd.Flags().GetString("aws-account")
	if err != nil {
		return nil, err
	}

	vpc, err := cmd.Flags().GetString("aws-vpc")
	if err != nil {
		return nil, err
	}

	routes, err := cmd.Flags().GetStringSlice("aws-routes")
	if err != nil {
		return nil, err
	}

	customerRegion, err := cmd.Flags().GetString("customer-region")
	if err != nil {
		return nil, err
	}
	if customerRegion == "" {
		customerRegion = networkRegion
	}

	awsPeering := &networkingv1.NetworkingV1AwsPeering{
		Kind:           "AwsPeering",
		Account:        account,
		Vpc:            vpc,
		Routes:         routes,
		CustomerRegion: customerRegion,
	}

	return awsPeering, nil
}

func (c *peeringCommand) createGcpPeeringRequest(cmd *cobra.Command) (*networkingv1.NetworkingV1GcpPeering, error) {
	project, err := cmd.Flags().GetString("gcp-project")
	if err != nil {
		return nil, err
	}

	vpcNetwork, err := cmd.Flags().GetString("gcp-vpc-network")
	if err != nil {
		return nil, err
	}

	importCustomRoutes, err := cmd.Flags().GetBool("gcp-import-custom-routes")
	if err != nil {
		return nil, err
	}

	gcpPeering := &networkingv1.NetworkingV1GcpPeering{
		Kind:               "GcpPeering",
		Project:            project,
		VpcNetwork:         vpcNetwork,
		ImportCustomRoutes: networkingv1.PtrBool(importCustomRoutes),
	}

	return gcpPeering, nil
}

func (c *peeringCommand) createAzurePeeringRequest(cmd *cobra.Command, networkRegion string) (*networkingv1.NetworkingV1AzurePeering, error) {
	tenant, err := cmd.Flags().GetString("azure-tenant")
	if err != nil {
		return nil, err
	}

	vnet, err := cmd.Flags().GetString("azure-vnet")
	if err != nil {
		return nil, err
	}

	customerRegion, err := cmd.Flags().GetString("customer-region")
	if err != nil {
		return nil, err
	}
	if customerRegion == "" {
		customerRegion = networkRegion
	}

	azurePeering := &networkingv1.NetworkingV1AzurePeering{
		Kind:           "AzurePeering",
		Tenant:         tenant,
		Vnet:           vnet,
		CustomerRegion: customerRegion,
	}

	return azurePeering, nil
}
