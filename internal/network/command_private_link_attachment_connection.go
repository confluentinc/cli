package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type privateLinkAttachmentConnectionOut struct {
	Id                           string `human:"ID" serialized:"id"`
	Name                         string `human:"Name,omitempty" serialized:"name,omitempty"`
	Cloud                        string `human:"Cloud" serialized:"cloud"`
	PrivateLinkAttachment        string `human:"Private Link Attachment" serialized:"private_link_attachment"`
	Phase                        string `human:"Phase" serialized:"phase"`
	AwsVpcEndpoint               string `human:"AWS VPC Endpoint,omitempty" serialized:"aws_vpc_endpoint,omitempty"`
	AwsVpcEndpointServiceName    string `human:"AWS VPC Endpoint Service Name,omitempty" serialized:"aws_vpc_endpoint_service_name,omitempty"`
	AzurePrivateEndpointResource string `human:"Azure Private Endpoint Resource,omitempty" serialized:"azure_private_endpoint_resource,omitempty"`
	AzurePrivateLinkServiceAlias string `human:"Azure Private Link Service Alias,omitempty" serialized:"azure_private_link_service_alias,omitempty"`
	AzurePrivateLinkService      string `human:"Azure Private Link Service,omitempty" serialized:"azure_private_link_service,omitempty"`
}

func (c *command) newPrivateLinkAttachmentConnectionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "Manage private link attachment connections.",
	}

	cmd.AddCommand(c.newPrivateLinkAttachmentConnectionCreateCommand())
	cmd.AddCommand(c.newPrivateLinkAttachmentConnectionDeleteCommand())
	cmd.AddCommand(c.newPrivateLinkAttachmentConnectionDescribeCommand())
	cmd.AddCommand(c.newPrivateLinkAttachmentConnectionListCommand())
	cmd.AddCommand(c.newPrivateLinkAttachmentConnectionUpdateCommand())

	return cmd
}

func printPrivateLinkAttachmentConnectionTable(cmd *cobra.Command, connection networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection) error {
	if connection.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if connection.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	out := &privateLinkAttachmentConnectionOut{
		Id:                    connection.GetId(),
		Name:                  connection.Spec.GetDisplayName(),
		PrivateLinkAttachment: connection.Spec.PrivateLinkAttachment.GetId(),
		Phase:                 connection.Status.GetPhase(),
	}

	if connection.Spec.HasCloud() {
		switch {
		case connection.Spec.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnection != nil:
			out.Cloud = CloudAws
			out.AwsVpcEndpoint = connection.Spec.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnection.GetVpcEndpointId()
		case connection.Spec.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnection != nil:
			out.Cloud = CloudAzure
			out.AzurePrivateEndpointResource = connection.Spec.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnection.GetPrivateEndpointResourceId()
		}
	}

	if connection.Status.HasCloud() {
		switch {
		case connection.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnectionStatus != nil:
			out.AwsVpcEndpointServiceName = connection.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnectionStatus.GetVpcEndpointServiceName()
		case connection.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnectionStatus != nil:
			out.AzurePrivateLinkServiceAlias = connection.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnectionStatus.GetPrivateLinkServiceAlias()
			out.AzurePrivateLinkService = connection.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnectionStatus.GetPrivateLinkServiceResourceId()
		}
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
