package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type transitGatewayAttachmentCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type transitGatewayAttachmentHumanOut struct {
	Id                            string `human:"ID"`
	Name                          string `human:"Name"`
	NetworkId                     string `human:"Network ID"`
	AwsRamShareArn                string `human:"AWS RAM Share ARN"`
	AwsTransitGatewayId           string `human:"AWS Transit Gateway ID"`
	Routes                        string `human:"Routes"`
	AwsTransitGatewayAttachmentId string `human:"AWS Transit Gateway Attachment ID,omitempty"`
	Phase                         string `human:"Phase"`
}

type transitGatewayAttachmentSerializedOut struct {
	Id                            string   `serialized:"id"`
	Name                          string   `serialized:"name"`
	NetworkId                     string   `serialized:"network_id"`
	AwsRamShareArn                string   `serialized:"aws_ram_share_arn"`
	AwsTransitGatewayId           string   `serialized:"aws_transit_gateway_id"`
	Routes                        []string `serialized:"routes"`
	AwsTransitGatewayAttachmentId string   `serialized:"aws_transit_gateway_attachment_id,omitempty"`
	Phase                         string   `serialized:"phase"`
}

func newTransitGatewayAttachmentCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transit-gateway-attachment",
		Aliases: []string{"tgwa"},
		Short:   "Manage transit gateway attachments.",
		Args:    cobra.NoArgs,
	}

	c := &transitGatewayAttachmentCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func (c *transitGatewayAttachmentCommand) getTransitGatewayAttachments() ([]networkingv1.NetworkingV1TransitGatewayAttachment, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListTransitGatewayAttachments(environmentId)
}

func (c *transitGatewayAttachmentCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validArgsMultiple(cmd, args)
}

func (c *transitGatewayAttachmentCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteTransitGatewayAttachments()
}

func (c *transitGatewayAttachmentCommand) autocompleteTransitGatewayAttachments() []string {
	attachments, err := c.getTransitGatewayAttachments()
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
	table := output.NewTable(cmd)

	if attachment.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if attachment.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	if output.GetFormat(cmd) == output.Human {
		table.Add(&transitGatewayAttachmentHumanOut{
			Id:                            attachment.GetId(),
			Name:                          attachment.Spec.GetDisplayName(),
			NetworkId:                     attachment.Spec.Network.GetId(),
			AwsRamShareArn:                attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
			AwsTransitGatewayId:           attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
			Routes:                        strings.Join(attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(), ", "),
			AwsTransitGatewayAttachmentId: attachment.Status.Cloud.NetworkingV1AwsTransitGatewayAttachmentStatus.GetTransitGatewayAttachmentId(),
			Phase:                         attachment.Status.GetPhase(),
		})
	} else {
		table.Add(&transitGatewayAttachmentSerializedOut{
			Id:                            attachment.GetId(),
			Name:                          attachment.Spec.GetDisplayName(),
			NetworkId:                     attachment.Spec.Network.GetId(),
			AwsRamShareArn:                attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
			AwsTransitGatewayId:           attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
			Routes:                        attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(),
			AwsTransitGatewayAttachmentId: attachment.Status.Cloud.NetworkingV1AwsTransitGatewayAttachmentStatus.GetTransitGatewayAttachmentId(),
			Phase:                         attachment.Status.GetPhase(),
		})
	}

	return table.PrintWithAutoWrap(false)
}
