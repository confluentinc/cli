package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type privateLinkAttachmentOut struct {
	Id                           string `human:"ID" serialized:"id"`
	Name                         string `human:"Name,omitempty" serialized:"name,omitempty"`
	Cloud                        string `human:"Cloud" serialized:"cloud"`
	Region                       string `human:"Region" serialized:"region"`
	AwsVpcEndpointService        string `human:"AWS VPC Endpoint Service,omitempty" serialized:"aws_vpc_endpoint_service,omitempty"`
	AzurePrivateLinkServiceAlias string `human:"Azure Private Link Service Alias,omitempty" serialized:"azure_private_link_service_alias,omitempty"`
	AzurePrivateLinkServiceId    string `human:"Azure Private Link Service Id,omitempty" serialized:"azure_private_link_service_id,omitempty"`
	Phase                        string `human:"Phase" serialized:"phase"`
}

func (c *command) newPrivateLinkAttachmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "Manage private link attachments.",
	}

	cmd.AddCommand(c.newPrivateLinkAttachmentConnectionCommand())
	cmd.AddCommand(c.newPrivateLinkAttachmentCreateCommand())
	cmd.AddCommand(c.newPrivateLinkAttachmentDeleteCommand())
	cmd.AddCommand(c.newPrivateLinkAttachmentDescribeCommand())
	cmd.AddCommand(c.newPrivateLinkAttachmentListCommand())
	cmd.AddCommand(c.newPrivateLinkAttachmentUpdateCommand())

	return cmd
}

func (c *command) getPrivateLinkAttachments() ([]networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListPrivateLinkAttachments(environmentId)
}

func (c *command) validPrivateLinkAttachmentArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validPrivateLinkAttachmentArgsMultiple(cmd, args)
}

func (c *command) validPrivateLinkAttachmentArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompletePrivateLinkAttachments()
}

func (c *command) autocompletePrivateLinkAttachments() []string {
	attachments, err := c.getPrivateLinkAttachments()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(attachments))
	for i, attachment := range attachments {
		suggestions[i] = fmt.Sprintf("%s\t%s", attachment.GetId(), attachment.Spec.GetDisplayName())
	}
	return suggestions
}

func printPrivateLinkAttachmentTable(cmd *cobra.Command, attachment networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment) error {
	if attachment.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if attachment.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	out := &privateLinkAttachmentOut{
		Id:     attachment.GetId(),
		Name:   attachment.Spec.GetDisplayName(),
		Cloud:  attachment.Spec.GetCloud(),
		Region: attachment.Spec.GetRegion(),
		Phase:  attachment.Status.GetPhase(),
	}

	if attachment.Status.Cloud != nil {
		switch {
		case attachment.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentStatus != nil:
			out.AwsVpcEndpointService = attachment.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentStatus.VpcEndpointService.GetVpcEndpointServiceName()
		case attachment.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentStatus != nil:
			out.AzurePrivateLinkServiceAlias = attachment.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentStatus.GetPrivateLinkService().PrivateLinkServiceAlias   // do we want to output id as well
			out.AzurePrivateLinkServiceId = attachment.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentStatus.GetPrivateLinkService().PrivateLinkServiceResourceId // is this necessary
		}

	}

	table := output.NewTable(cmd)
	table.Add(out)

	return table.Print()
}
