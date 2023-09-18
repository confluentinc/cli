package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newPrivateLinkAttachmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List private link attachments.",
		Args:  cobra.NoArgs,
		RunE:  c.privateLinkAttachmentList,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) privateLinkAttachmentList(cmd *cobra.Command, _ []string) error {
	attachments, err := c.getPrivateLinkAttachments()
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

		out := &privateLinkAttachmentOut{
			Id:                    attachment.GetId(),
			Name:                  attachment.Spec.GetDisplayName(),
			Cloud:                 attachment.Spec.GetCloud(),
			Region:                attachment.Spec.GetRegion(),
			AwsVpcEndpointService: attachment.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentStatus.VpcEndpointService.GetVpcEndpointServiceName(),
			Phase:                 attachment.Status.GetPhase(),
		}

		list.Add(out)
	}

	return list.Print()
}
