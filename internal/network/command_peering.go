package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type peeringCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type peeringHumanOut struct {
	Id            string `human:"ID"`
	Name          string `human:"Name"`
	NetworkId     string `human:"Network ID"`
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
	Name          string   `serialized:"name"`
	NetworkId     string   `serialized:"network_id"`
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

func newPeeringCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "peering",
		Short:       "Manage peering connections.",
		Args:        cobra.NoArgs,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &peeringCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newPeeringListCommand())

	return cmd
}

func (c *peeringCommand) getPeerings() ([]networkingv1.NetworkingV1Peering, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListPeerings(environmentId)
}
