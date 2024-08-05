package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type transitGatewayAttachmentOut struct {
	Id                            string   `human:"ID" serialized:"id"`
	Name                          string   `human:"Name,omitempty" serialized:"name,omitempty"`
	Network                       string   `human:"Network" serialized:"network"`
	AwsRamShareArn                string   `human:"AWS RAM Share ARN" serialized:"aws_ram_share_arn"`
	AwsTransitGatewayId           string   `human:"AWS Transit Gateway ID" serialized:"aws_transit_gateway_id"`
	Routes                        []string `human:"Routes" serialized:"routes"`
	AwsTransitGatewayAttachmentId string   `human:"AWS Transit Gateway Attachment ID,omitempty" serialized:"aws_transit_gateway_attachment_id,omitempty"`
	Phase                         string   `human:"Phase" serialized:"phase"`
}

func (c *command) newTransitGatewayAttachmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transit-gateway-attachment",
		Aliases: []string{"tgwa"},
		Short:   "Manage transit gateway attachments.",
		Args:    cobra.NoArgs,
	}

	cmd.AddCommand(c.newTransitGatewayAttachmentCreateCommand())
	cmd.AddCommand(c.newTransitGatewayAttachmentDeleteCommand())
	cmd.AddCommand(c.newTransitGatewayAttachmentDescribeCommand())
	cmd.AddCommand(c.newTransitGatewayAttachmentListCommand())
	cmd.AddCommand(c.newTransitGatewayAttachmentUpdateCommand())

	return cmd
}

func (c *command) getTransitGatewayAttachments(name, network, phase []string) ([]networkingv1.NetworkingV1TransitGatewayAttachment, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListTransitGatewayAttachments(environmentId, name, network, phase)
}

func (c *command) validTransitGatewayAttachmentArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validTransitGatewayAttachmentArgsMultiple(cmd, args)
}

func (c *command) validTransitGatewayAttachmentArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteTransitGatewayAttachments()
}

func (c *command) autocompleteTransitGatewayAttachments() []string {
	attachments, err := c.getTransitGatewayAttachments(nil, nil, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(attachments))
	for i, attachment := range attachments {
		suggestions[i] = fmt.Sprintf("%s\t%s", attachment.GetId(), attachment.Spec.GetDisplayName())
	}
	return suggestions
}

func printTransitGatewayAttachmentTable(cmd *cobra.Command, attachment networkingv1.NetworkingV1TransitGatewayAttachment) error {
	if attachment.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if attachment.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	table := output.NewTable(cmd)

	table.Add(&transitGatewayAttachmentOut{
		Id:                            attachment.GetId(),
		Name:                          attachment.Spec.GetDisplayName(),
		Network:                       attachment.Spec.Network.GetId(),
		AwsRamShareArn:                attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
		AwsTransitGatewayId:           attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
		Routes:                        attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(),
		AwsTransitGatewayAttachmentId: attachment.Status.Cloud.NetworkingV1AwsTransitGatewayAttachmentStatus.GetTransitGatewayAttachmentId(),
		Phase:                         attachment.Status.GetPhase(),
	})
	return table.PrintWithAutoWrap(false)
}
