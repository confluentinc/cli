package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newPrivateLinkAttachmentConnectionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "list <private-link-attachment-id>",
		Short:             "List connections for a private link attachment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPrivateLinkAttachmentArgs),
		RunE:              c.privateLinkAttachmentConnectionList,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) privateLinkAttachmentConnectionList(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	connections, err := c.V2Client.ListPrivateLinkAttachmentConnections(environmentId, args[0])
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, connection := range connections {
		if connection.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if connection.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		out := &privateLinkAttachmentConnectionOut{
			Id:                      connection.GetId(),
			Name:                    connection.Spec.GetDisplayName(),
			PrivateLinkAttachmentId: connection.Spec.PrivateLinkAttachment.GetId(),
			Phase:                   connection.Status.GetPhase(),
		}

		if connection.Spec.Cloud != nil && connection.Spec.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnection != nil {
			out.Cloud = CloudAws
			out.AwsVpcEndpointId = connection.Spec.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnection.GetVpcEndpointId()
		}

		if connection.Status.Cloud != nil && connection.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnectionStatus != nil {
			out.AwsVpcEndpointServiceName = connection.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnectionStatus.GetVpcEndpointServiceName()
		}

		list.Add(out)
	}

	return list.Print()
}
