package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newPrivateLinkAttachmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List private link attachments.",
		Args:  cobra.NoArgs,
		RunE:  c.privateLinkAttachmentList,
	}

	cmd.Flags().StringSlice("name", nil, "A comma-separated list of private link attachment names.")
	pcmd.AddListCloudFlag(cmd)
	c.addListRegionFlagNetwork(cmd, c.AuthenticatedCLICommand)
	addPhaseFlag(cmd, resource.PrivateLinkAttachment)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) privateLinkAttachmentList(cmd *cobra.Command, _ []string) error {
	name, err := cmd.Flags().GetStringSlice("name")
	if err != nil {
		return err
	}

	cloud, err := cmd.Flags().GetStringSlice("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetStringSlice("region")
	if err != nil {
		return err
	}

	phase, err := cmd.Flags().GetStringSlice("phase")
	if err != nil {
		return err
	}

	cloud, phase = toUpper(cloud), toUpper(phase)

	attachments, err := c.getPrivateLinkAttachments(name, cloud, region, phase)
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
			Id:     attachment.GetId(),
			Name:   attachment.Spec.GetDisplayName(),
			Cloud:  attachment.Spec.GetCloud(),
			Region: attachment.Spec.GetRegion(),
			Phase:  attachment.Status.GetPhase(),
		}

		if attachment.Status.Cloud != nil && attachment.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentStatus != nil {
			out.AwsVpcEndpointService = attachment.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentStatus.VpcEndpointService.GetVpcEndpointServiceName()
		}

		list.Add(out)
	}

	return list.Print()
}
