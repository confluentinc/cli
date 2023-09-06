package network

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *transitGatewayAttachmentCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transit gateway attachments.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *transitGatewayAttachmentCommand) list(cmd *cobra.Command, _ []string) error {
	tgwas, err := c.getTransitGatewayAttachments()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, tgwa := range tgwas {
		if tgwa.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if tgwa.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		if output.GetFormat(cmd) == output.Human {
			list.Add(&transitGatewayAttachmentHumanOut{
				Id:                         tgwa.GetId(),
				Name:                       tgwa.Spec.GetDisplayName(),
				NetworkId:                  tgwa.Spec.Network.GetId(),
				RamShareArn:                tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
				TransitGatewayId:           tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
				Routes:                     strings.Join(tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(), ", "),
				TransitGatewayAttachmentId: tgwa.Status.Cloud.NetworkingV1AwsTransitGatewayAttachmentStatus.GetTransitGatewayAttachmentId(),
				Phase:                      tgwa.Status.GetPhase(),
			})
		} else {
			list.Add(&transitGatewayAttachmentSerializedOut{
				Id:                         tgwa.GetId(),
				Name:                       tgwa.Spec.GetDisplayName(),
				NetworkId:                  tgwa.Spec.Network.GetId(),
				RamShareArn:                tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRamShareArn(),
				TransitGatewayId:           tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetTransitGatewayId(),
				Routes:                     tgwa.Spec.Cloud.NetworkingV1AwsTransitGatewayAttachment.GetRoutes(),
				TransitGatewayAttachmentId: tgwa.Status.Cloud.NetworkingV1AwsTransitGatewayAttachmentStatus.GetTransitGatewayAttachmentId(),
				Phase:                      tgwa.Status.GetPhase(),
			})
		}
	}
	return list.Print()
}
