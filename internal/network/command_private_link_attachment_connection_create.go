package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"

	pcloud "github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newPrivateLinkAttachmentConnectionCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a private link attachment connection.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.privateLinkAttachmentConnectionCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create a Private Link attachment connection named "aws-private-link-attachment-connection".`,
				Code: "confluent network private-link attachment connection create aws-private-link-attachment-connection --cloud aws --endpoint vpce-1234567890abcdef0 --attachment platt-123456",
			},
			examples.Example{
				Text: `Create a Private Link attachment connection named "gcp-private-link-attachment-connection".`,
				Code: "confluent network private-link attachment connection create gcp-private-link-attachment-connection --cloud gcp --endpoint 1234567890123456 --attachment platt-123456",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("endpoint", "", "ID of an endpoint that is connected to either AWS VPC endpoint service, Azure PrivateLink service, or GCP Private Service Connect service.")
	c.addPrivateLinkAttachmentFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("endpoint"))
	cobra.CheckErr(cmd.MarkFlagRequired("attachment"))

	return cmd
}

func (c *command) privateLinkAttachmentConnectionCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)

	endpoint, err := cmd.Flags().GetString("endpoint")
	if err != nil {
		return err
	}

	attachment, err := cmd.Flags().GetString("attachment")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	createPrivateLinkAttachmentConnection := networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnection{
		Spec: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionSpec{
			Environment:           &networkingprivatelinkv1.ObjectReference{Id: environmentId},
			PrivateLinkAttachment: &networkingprivatelinkv1.ObjectReference{Id: attachment},
		},
	}

	if name != "" {
		createPrivateLinkAttachmentConnection.Spec.SetDisplayName(name)
	}

	switch cloud {
	case pcloud.Aws:
		createPrivateLinkAttachmentConnection.Spec.Cloud = &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionSpecCloudOneOf{
			NetworkingV1AwsPrivateLinkAttachmentConnection: &networkingprivatelinkv1.NetworkingV1AwsPrivateLinkAttachmentConnection{
				Kind:          "AwsPrivateLinkAttachmentConnection",
				VpcEndpointId: endpoint,
			},
		}
	case pcloud.Gcp:
		createPrivateLinkAttachmentConnection.Spec.Cloud = &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionSpecCloudOneOf{
			NetworkingV1GcpPrivateLinkAttachmentConnection: &networkingprivatelinkv1.NetworkingV1GcpPrivateLinkAttachmentConnection{
				Kind:                              "GcpPrivateLinkAttachmentConnection",
				PrivateServiceConnectConnectionId: endpoint,
			},
		}
	case pcloud.Azure:
		createPrivateLinkAttachmentConnection.Spec.Cloud = &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionSpecCloudOneOf{
			NetworkingV1AzurePrivateLinkAttachmentConnection: &networkingprivatelinkv1.NetworkingV1AzurePrivateLinkAttachmentConnection{
				Kind:                      "AzurePrivateLinkAttachmentConnection",
				PrivateEndpointResourceId: endpoint,
			},
		}
	}

	connection, err := c.V2Client.CreatePrivateLinkAttachmentConnection(createPrivateLinkAttachmentConnection)
	if err != nil {
		return err
	}

	return printPrivateLinkAttachmentConnectionTable(cmd, connection)
}
