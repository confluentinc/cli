package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcloud "github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newPrivateLinkAttachmentConnectionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connections for a private link attachment.",
		Args:  cobra.NoArgs,
		RunE:  c.privateLinkAttachmentConnectionList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List connections for private link attachment "platt-123456".`,
				Code: "confluent network private-link attachment connection list --attachment platt-123456",
			},
		),
	}

	c.addPrivateLinkAttachmentFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("attachment"))

	return cmd
}

func (c *command) privateLinkAttachmentConnectionList(cmd *cobra.Command, _ []string) error {
	attachment, err := cmd.Flags().GetString("attachment")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	connections, err := c.V2Client.ListPrivateLinkAttachmentConnections(environmentId, attachment)
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
			Id:                    connection.GetId(),
			Name:                  connection.Spec.GetDisplayName(),
			PrivateLinkAttachment: connection.Spec.PrivateLinkAttachment.GetId(),
			Phase:                 connection.Status.GetPhase(),
		}

		if connection.Spec.HasCloud() {
			switch {
			case connection.Spec.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnection != nil:
				out.Cloud = pcloud.Aws
				out.AwsVpcEndpointId = connection.Spec.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnection.GetVpcEndpointId()
			case connection.Spec.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnection != nil:
				out.Cloud = pcloud.Azure
				out.AzurePrivateEndpointResourceId = connection.Spec.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnection.GetPrivateEndpointResourceId()
			case connection.Spec.Cloud.NetworkingV1GcpPrivateLinkAttachmentConnection != nil:
				out.Cloud = pcloud.Gcp
				out.GcpPrivateServiceConnectConnectionId = connection.Spec.Cloud.NetworkingV1GcpPrivateLinkAttachmentConnection.GetPrivateServiceConnectConnectionId()
			}
		}

		if connection.Status.HasCloud() {
			switch {
			case connection.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnectionStatus != nil:
				out.AwsVpcEndpointServiceName = connection.Status.Cloud.NetworkingV1AwsPrivateLinkAttachmentConnectionStatus.GetVpcEndpointServiceName()
			case connection.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnectionStatus != nil:
				out.AzurePrivateLinkServiceAlias = connection.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnectionStatus.GetPrivateLinkServiceAlias()
				out.AzurePrivateLinkServiceId = connection.Status.Cloud.NetworkingV1AzurePrivateLinkAttachmentConnectionStatus.GetPrivateLinkServiceResourceId()
			case connection.Status.Cloud.NetworkingV1GcpPrivateLinkAttachmentConnectionStatus != nil:
				out.GcpServiceAttachmentId = connection.Status.Cloud.NetworkingV1GcpPrivateLinkAttachmentConnectionStatus.GetPrivateServiceConnectServiceAttachment()
			}
		}

		list.Add(out)
	}

	return list.Print()
}
