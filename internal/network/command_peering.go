package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type peeringHumanOut struct {
	Id            string `human:"ID"`
	Name          string `human:"Name,omitempty"`
	Network       string `human:"Network"`
	Cloud         string `human:"Cloud"`
	Phase         string `human:"Phase"`
	CustomRegion  string `human:"Custom Region,omitempty"`
	AwsVpc        string `human:"AWS VPC,omitempty"`
	AwsAccount    string `human:"AWS Account,omitempty"`
	AwsRoutes     string `human:"AWS Routes,omitempty"`
	GcpProject    string `human:"GCP Project,omitempty"`
	GcpVpcNetwork string `human:"GCP VPC Network,omitempty"`
	AzureVNet     string `human:"Azure VNet,omitempty"`
	AzureTenant   string `human:"Azure Tenant,omitempty"`
}

type peeringSerializedOut struct {
	Id            string   `serialized:"id"`
	Name          string   `serialized:"name,omitempty"`
	Network       string   `serialized:"network"`
	Cloud         string   `serialized:"cloud"`
	Phase         string   `serialized:"phase"`
	CustomRegion  string   `serialized:"custom_region,omitempty"`
	AwsVpc        string   `serialized:"aws_vpc,omitempty"`
	AwsAccount    string   `serialized:"aws_account,omitempty"`
	AwsRoutes     []string `serialized:"aws_routes,omitempty"`
	GcpProject    string   `serialized:"gcp_project,omitempty"`
	GcpVpcNetwork string   `serialized:"gcp_vpc_network,omitempty"`
	AzureVNet     string   `serialized:"azure_vnet,omitempty"`
	AzureTenant   string   `serialized:"azure_tenant,omitempty"`
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

func (c *command) getPeerings() ([]networkingv1.NetworkingV1Peering, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListPeerings(environmentId)
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
	peerings, err := c.getPeerings()
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
		return CloudAws, nil
	} else if cloud.NetworkingV1GcpPeering != nil {
		return CloudGcp, nil
	} else if cloud.NetworkingV1AzurePeering != nil {
		return CloudAzure, nil
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

	human := &peeringHumanOut{
		Id:      peering.GetId(),
		Name:    peering.Spec.GetDisplayName(),
		Network: peering.Spec.Network.GetId(),
		Cloud:   cloud,
		Phase:   peering.Status.GetPhase(),
	}

	serialized := &peeringSerializedOut{
		Id:      peering.GetId(),
		Name:    peering.Spec.GetDisplayName(),
		Network: peering.Spec.Network.GetId(),
		Cloud:   cloud,
		Phase:   peering.Status.GetPhase(),
	}

	describeFields := []string{"Id", "Name", "Network", "Cloud", "Phase"}

	switch cloud {
	case CloudAws:
		human.AwsVpc = peering.Spec.Cloud.NetworkingV1AwsPeering.GetVpc()
		human.AwsAccount = peering.Spec.Cloud.NetworkingV1AwsPeering.GetAccount()
		human.AwsRoutes = strings.Join(peering.Spec.Cloud.NetworkingV1AwsPeering.GetRoutes(), ", ")
		human.CustomRegion = peering.Spec.Cloud.NetworkingV1AwsPeering.GetCustomerRegion()
		serialized.AwsVpc = peering.Spec.Cloud.NetworkingV1AwsPeering.GetVpc()
		serialized.AwsAccount = peering.Spec.Cloud.NetworkingV1AwsPeering.GetAccount()
		serialized.AwsRoutes = peering.Spec.Cloud.NetworkingV1AwsPeering.GetRoutes()
		serialized.CustomRegion = peering.Spec.Cloud.NetworkingV1AwsPeering.GetCustomerRegion()
		describeFields = append(describeFields, "AwsVpc", "AwsAccount", "AwsRoutes", "CustomRegion")
	case CloudGcp:
		human.GcpVpcNetwork = peering.Spec.Cloud.NetworkingV1GcpPeering.GetVpcNetwork()
		human.GcpProject = peering.Spec.Cloud.NetworkingV1GcpPeering.GetProject()
		serialized.GcpVpcNetwork = peering.Spec.Cloud.NetworkingV1GcpPeering.GetVpcNetwork()
		serialized.GcpProject = peering.Spec.Cloud.NetworkingV1GcpPeering.GetProject()
		describeFields = append(describeFields, "GcpVpcNetwork", "GcpProject")
	case CloudAzure:
		human.AzureVNet = peering.Spec.Cloud.NetworkingV1AzurePeering.GetVnet()
		human.AzureTenant = peering.Spec.Cloud.NetworkingV1AzurePeering.GetTenant()
		human.CustomRegion = peering.Spec.Cloud.NetworkingV1AzurePeering.GetCustomerRegion()
		serialized.AzureVNet = peering.Spec.Cloud.NetworkingV1AzurePeering.GetVnet()
		serialized.AzureTenant = peering.Spec.Cloud.NetworkingV1AzurePeering.GetTenant()
		serialized.CustomRegion = peering.Spec.Cloud.NetworkingV1AzurePeering.GetCustomerRegion()
		describeFields = append(describeFields, "AzureVNet", "AzureTenant", "CustomRegion")
	}

	table := output.NewTable(cmd)

	if output.GetFormat(cmd) == output.Human {
		table.Add(human)
	} else {
		table.Add(serialized)
	}

	table.Filter(describeFields)
	return table.Print()
}
