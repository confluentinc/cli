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
	Id                         string `human:"ID"`
	Name                       string `human:"Name"`
	NetworkId                  string `human:"Network ID"`
	RamShareArn                string `human:"RAM Share ARN"`
	TransitGatewayId           string `human:"Transit Gateway ID"`
	Routes                     string `human:"Routes"`
	TransitGatewayAttachmentId string `human:"Transit Gateway Attachment ID,omitempty"`
	Phase                      string `human:"Phase"`
}

type transitGatewayAttachmentSerializedOut struct {
	Id                         string   `serialized:"id"`
	Name                       string   `serialized:"name"`
	NetworkId                  string   `serialized:"network_id"`
	RamShareArn                string   `serialized:"ram_share_arn"`
	TransitGatewayId           string   `serialized:"transit_gateway_id"`
	Routes                     []string `serialized:"routes"`
	TransitGatewayAttachmentId string   `serialized:"transit_gateway_attachment_id,omitempty"`
	Phase                      string   `serialized:"phase"`
}

func newTransitGatewayAttachmentCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transit-gateway-attachment",
		Aliases: []string{"tgw-attachment"},
		Short:   "Manage transit gateway attachments.",
		Args:    cobra.NoArgs,
	}

	c := &transitGatewayAttachmentCommand{AuthenticatedCLICommand: pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())

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
	tgwas, err := c.getTransitGatewayAttachments()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(tgwas))
	for i, tgwa := range tgwas {
		suggestions[i] = fmt.Sprintf("%s\t%s", tgwa.GetId(), tgwa.Spec.GetDisplayName())
	}
	return suggestions
}

func printTransitGatewayAttachmentTable(cmd *cobra.Command, tgwa networkingv1.NetworkingV1TransitGatewayAttachment) error {
	table := output.NewTable(cmd)

	if tgwa.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if tgwa.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	human := &transitGatewayAttachmentHumanOut{
		Id:                         tgwa.GetId(),
		Name:                       tgwa.Spec.GetDisplayName(),
		NetworkId:                  tgwa.Spec.Network.GetId(),
		RamShareArn:                tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
		TransitGatewayId:           tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
		Routes:                     strings.Join(tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(), ", "),
		TransitGatewayAttachmentId: tgwa.Status.Cloud.NetworkingV1AwsTransitGatewayAttachmentStatus.GetTransitGatewayAttachmentId(),
		Phase:                      tgwa.Status.GetPhase(),
	}

	serialized := &transitGatewayAttachmentSerializedOut{
		Id:                         tgwa.GetId(),
		Name:                       tgwa.Spec.GetDisplayName(),
		NetworkId:                  tgwa.Spec.Network.GetId(),
		RamShareArn:                tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
		TransitGatewayId:           tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
		Routes:                     tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(),
		TransitGatewayAttachmentId: tgwa.Status.Cloud.NetworkingV1AwsTransitGatewayAttachmentStatus.GetTransitGatewayAttachmentId(),
		Phase:                      tgwa.Status.GetPhase(),
	}

	if output.GetFormat(cmd) == output.Human {
		table.Add(human)
	} else {
		table.Add(serialized)
	}

	return table.Print()
}
