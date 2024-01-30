package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newTransitGatewayAttachmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transit gateway attachments.",
		Args:  cobra.NoArgs,
		RunE:  c.transitGatewayAttachmentList,
	}
	cmd.Flags().StringSlice("name", nil, "A comma-separated list of transit gateway attachment names.")
	addListNetworkFlag(cmd, c.AuthenticatedCLICommand)
	addPhaseFlag(cmd, resource.TransitGatewayAttachment)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) transitGatewayAttachmentList(cmd *cobra.Command, _ []string) error {
	name, err := cmd.Flags().GetStringSlice("name")
	if err != nil {
		return err
	}

	network, err := cmd.Flags().GetStringSlice("network")
	if err != nil {
		return err
	}

	phase, err := cmd.Flags().GetStringSlice("phase")
	if err != nil {
		return err
	}

	phase = toUpper(phase)

	attachments, err := c.getTransitGatewayAttachments(name, network, phase)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, attachment := range attachments {
		if attachment.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if attachment.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		if output.GetFormat(cmd) == output.Human {
			list.Add(&transitGatewayAttachmentHumanOut{
				Id:                            attachment.GetId(),
				Name:                          attachment.Spec.GetDisplayName(),
				Network:                       attachment.Spec.Network.GetId(),
				AwsRamShareArn:                attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
				AwsTransitGatewayId:           attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
				Routes:                        strings.Join(attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(), ", "),
				AwsTransitGatewayAttachmentId: attachment.Status.Cloud.NetworkingV1AwsTransitGatewayAttachmentStatus.GetTransitGatewayAttachmentId(),
				Phase:                         attachment.Status.GetPhase(),
			})
		} else {
			list.Add(&transitGatewayAttachmentSerializedOut{
				Id:                            attachment.GetId(),
				Name:                          attachment.Spec.GetDisplayName(),
				Network:                       attachment.Spec.Network.GetId(),
				AwsRamShareArn:                attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
				AwsTransitGatewayId:           attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
				Routes:                        attachment.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(),
				AwsTransitGatewayAttachmentId: attachment.Status.Cloud.NetworkingV1AwsTransitGatewayAttachmentStatus.GetTransitGatewayAttachmentId(),
				Phase:                         attachment.Status.GetPhase(),
			})
		}
	}
	return list.Print()
}
