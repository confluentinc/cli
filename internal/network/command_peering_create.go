package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPeeringCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a peering.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.peeringCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an AWS VPC peering.",
				Code: "confluent network peering create aws-peering --network n-123456 --cloud aws --cloud-account 123456789012 --virtual-network vpc-1234567890abcdef0 --aws-routes 172.31.0.0/16,10.108.16.0/21",
			},
			examples.Example{
				Text: "Create a GCP VPC peering.",
				Code: "confluent network peering create gcp-peering --network n-123456 --cloud gcp --cloud-account temp-123456 --virtual-network customer-test-vpc-network --gcp-routes",
			},
			examples.Example{
				Text: "Create an Azure VNet peering.",
				Code: "confluent network peering create azure-peering --network n-123456 --cloud azure --cloud-account 1111tttt-1111-1111-1111-111111tttttt --virtual-network /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet --customer-region centralus",
			},
		),
	}

	addNetworkFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("cloud-account", "", "AWS account ID or Google Cloud project ID associated with the VPC that you are peering with Confluent Cloud network or Azure Tenant ID in which your Azure Subscription exists.")
	cmd.Flags().String("virtual-network", "", "AWS VPC ID, name of the Google Cloud VPC, or Azure Resource ID of the VNet that you are peering with Confluent Cloud network.")
	cmd.Flags().String("customer-region", "", "Cloud region ID of the AWS VPC or Azure VNet that you are peering with Confluent Cloud network.")
	cmd.Flags().StringSlice("aws-routes", nil, "A comma-separated list of CIDR blocks of the AWS VPC that you are peering with Confluent Cloud network. Required for AWS VPC Peering.")
	cmd.Flags().Bool("gcp-routes", false, "Enable customer route import for Google Cloud VPC Peering.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("network"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud-account"))
	cobra.CheckErr(cmd.MarkFlagRequired("virtual-network"))

	return cmd
}

func (c *command) peeringCreate(cmd *cobra.Command, args []string) error {
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
			DisplayName: networkingv1.PtrString(name),
			Environment: &networkingv1.ObjectReference{Id: environmentId},
			Network:     &networkingv1.ObjectReference{Id: networkId},
		},
	}

	switch cloud {
	case CloudAws:
		awsPeering, err := createAwsPeeringRequest(cmd, region)
		if err != nil {
			return err
		}
		createPeering.Spec.Cloud = &networkingv1.NetworkingV1PeeringSpecCloudOneOf{
			NetworkingV1AwsPeering: awsPeering,
		}
	case CloudGcp:
		gcpPeering, err := createGcpPeeringRequest(cmd)
		if err != nil {
			return err
		}
		createPeering.Spec.Cloud = &networkingv1.NetworkingV1PeeringSpecCloudOneOf{
			NetworkingV1GcpPeering: gcpPeering,
		}
	case CloudAzure:
		azurePeering, err := createAzurePeeringRequest(cmd, region)
		if err != nil {
			return err
		}
		createPeering.Spec.Cloud = &networkingv1.NetworkingV1PeeringSpecCloudOneOf{
			NetworkingV1AzurePeering: azurePeering,
		}
	}

	peering, err := c.V2Client.CreatePeering(createPeering)
	if err != nil {
		return err
	}

	return printPeeringTable(cmd, peering)
}

func createAwsPeeringRequest(cmd *cobra.Command, networkRegion string) (*networkingv1.NetworkingV1AwsPeering, error) {
	account, err := cmd.Flags().GetString("cloud-account")
	if err != nil {
		return nil, err
	}

	vpc, err := cmd.Flags().GetString("virtual-network")
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

func createGcpPeeringRequest(cmd *cobra.Command) (*networkingv1.NetworkingV1GcpPeering, error) {
	project, err := cmd.Flags().GetString("cloud-account")
	if err != nil {
		return nil, err
	}

	vpcNetwork, err := cmd.Flags().GetString("virtual-network")
	if err != nil {
		return nil, err
	}

	gcpImportCustomRoutes, err := cmd.Flags().GetBool("gcp-routes")
	if err != nil {
		return nil, err
	}

	gcpPeering := &networkingv1.NetworkingV1GcpPeering{
		Kind:               "GcpPeering",
		Project:            project,
		VpcNetwork:         vpcNetwork,
		ImportCustomRoutes: networkingv1.PtrBool(gcpImportCustomRoutes),
	}

	return gcpPeering, nil
}

func createAzurePeeringRequest(cmd *cobra.Command, networkRegion string) (*networkingv1.NetworkingV1AzurePeering, error) {
	tenant, err := cmd.Flags().GetString("cloud-account")
	if err != nil {
		return nil, err
	}

	vnet, err := cmd.Flags().GetString("virtual-network")
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
