package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
)

type accessPointCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type accessPointOut struct {
	Id                              string `human:"ID" serialized:"id"`
	Name                            string `human:"Name,omitempty" serialized:"name,omitempty"`
	AwsVpcEndpointService           string `human:"AWS VPC Endpoint Service,omitempty" serialized:"aws_vpc_endpoint_service,omitempty"`
	AwsVpcEndpoint                  string `human:"AWS VPC Endpoint,omitempty" serialized:"aws_vpc_endpoint,omitempty"`
	AwsVpcEndpointDnsName           string `human:"AWS VPC Endpoint DNS Name,omitempty" serialized:"aws_vpc_endpoint_dns_name,omitempty"`
	AzurePrivateLinkService         string `human:"Azure Private Link Service,omitempty" serialized:"azure_private_link_service,omitempty"`
	AzurePrivateLinkSubresourceName string `human:"Azure Private Link Subresource Name,omitempty" serialized:"azure_private_link_subresource_name,omitempty"`
	AzurePrivateEndpoint            string `human:"Azure Private Endpoint,omitempty" serialized:"azure_private_endpoint,omitempty"`
	AzurePrivateEndpointDomain      string `human:"Azure Private Endpoint Domain,omitempty" serialized:"azure_private_endpoint_domain,omitempty"`
	AzurePrivateEndpointIpAddress   string `human:"Azure Private Endpoint IP Address,omitempty" serialized:"azure_private_endpoint_ip_address,omitempty"`
	HighAvailability                bool   `human:"High Availability,omitempty" serialized:"high_availability,omitempty"`
	Environment                     string `human:"Environment" serialized:"environment"`
	Gateway                         string `human:"Gateway" serialized:"gateway"`
	Phase                           string `human:"Phase" serialized:"phase"`
}

func newAccessPointCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "access-point",
		Short:       "Manage access points.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &accessPointCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newPrivateLinkCommand())

	return cmd
}
