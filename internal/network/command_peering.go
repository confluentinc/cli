package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

type peeringOut struct {
	Id            string   `human:"ID" serialized:"id"`
	Name          string   `human:"Name,omitempty" serialized:"name,omitempty"`
	Network       string   `human:"Network" serialized:"network"`
	Cloud         string   `human:"Cloud" serialized:"cloud"`
	Phase         string   `human:"Phase" serialized:"phase"`
	CustomRegion  string   `human:"Custom Region,omitempty" serialized:"custom_region,omitempty"`
	AwsVpc        string   `human:"AWS VPC,omitempty" serialized:"aws_vpc,omitempty"`
	AwsAccount    string   `human:"AWS Account,omitempty" serialized:"aws_account,omitempty"`
	AwsRoutes     []string `human:"AWS Routes,omitempty" serialized:"aws_routes,omitempty"`
	GcpProject    string   `human:"GCP Project,omitempty" serialized:"gcp_project,omitempty"`
	GcpVpcNetwork string   `human:"GCP VPC Network,omitempty" serialized:"gcp_vpc_network,omitempty"`
	AzureVNet     string   `human:"Azure VNet,omitempty" serialized:"azure_vnet,omitempty"`
	AzureTenant   string   `human:"Azure Tenant,omitempty" serialized:"azure_tenant,omitempty"`
}

func (c *command) newPeeringCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "peering",
		Short: "Manage peerings.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newPeeringCreateCommand())
	cmd.AddCommand(c.newPeeringDeleteCommand())
	cmd.AddCommand(c.newPeeringDescribeCommand())
	cmd.AddCommand(c.newPeeringListCommand())
	cmd.AddCommand(c.newPeeringUpdateCommand())

	return cmd
}

func (c *command) getPeerings(name, network, phase []string) ([]networkingv1.NetworkingV1Peering, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListPeerings(environmentId, name, network, phase)
}

func (c *command) validPeeringArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validPeeringArgsMultiple(cmd, args)
}

func (c *command) validPeeringArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompletePeerings()
}

func (c *command) autocompletePeerings() []string {
	peerings, err := c.getPeerings(nil, nil, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(peerings))
	for i, peering := range peerings {
		suggestions[i] = fmt.Sprintf("%s\t%s", peering.GetId(), peering.Spec.GetDisplayName())
	}
	return suggestions
}

func getPeeringCloud(peering networkingv1.NetworkingV1Peering) (string, error) {
	cloud := peering.Spec.GetCloud()

	if cloud.NetworkingV1AwsPeering != nil {
		return resource.CloudAws, nil
	} else if cloud.NetworkingV1GcpPeering != nil {
		return resource.CloudGcp, nil
	} else if cloud.NetworkingV1AzurePeering != nil {
		return resource.CloudAzure, nil
	}

	return "", fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
}

func printPeeringTable(cmd *cobra.Command, peering networkingv1.NetworkingV1Peering) error {
	if peering.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if peering.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	cloud, err := getPeeringCloud(peering)
	if err != nil {
		return err
	}

	out := &peeringOut{
		Id:      peering.GetId(),
		Name:    peering.Spec.GetDisplayName(),
		Network: peering.Spec.Network.GetId(),
		Cloud:   cloud,
		Phase:   peering.Status.GetPhase(),
	}

	describeFields := []string{"Id", "Name", "Network", "Cloud", "Phase"}

	switch cloud {
	case resource.CloudAws:
		out.AwsVpc = peering.Spec.Cloud.NetworkingV1AwsPeering.GetVpc()
		out.AwsAccount = peering.Spec.Cloud.NetworkingV1AwsPeering.GetAccount()
		out.AwsRoutes = peering.Spec.Cloud.NetworkingV1AwsPeering.GetRoutes()
		out.CustomRegion = peering.Spec.Cloud.NetworkingV1AwsPeering.GetCustomerRegion()
		describeFields = append(describeFields, "AwsVpc", "AwsAccount", "AwsRoutes", "CustomRegion")
	case resource.CloudGcp:
		out.GcpVpcNetwork = peering.Spec.Cloud.NetworkingV1GcpPeering.GetVpcNetwork()
		out.GcpProject = peering.Spec.Cloud.NetworkingV1GcpPeering.GetProject()
		describeFields = append(describeFields, "GcpVpcNetwork", "GcpProject")
	case resource.CloudAzure:
		out.AzureVNet = peering.Spec.Cloud.NetworkingV1AzurePeering.GetVnet()
		out.AzureTenant = peering.Spec.Cloud.NetworkingV1AzurePeering.GetTenant()
		out.CustomRegion = peering.Spec.Cloud.NetworkingV1AzurePeering.GetCustomerRegion()
		describeFields = append(describeFields, "AzureVNet", "AzureTenant", "CustomRegion")
	}

	table := output.NewTable(cmd)
	table.Add(out)
	table.Filter(describeFields)
	return table.Print()
}
