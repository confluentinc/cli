package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPrivateLinkAttachmentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a private link attachment.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.privateLinkAttachmentCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an AWS PrivateLink attachment with a display name.",
				Code: "confluent network private-link attachment create aws-private-link-attachment --cloud aws --region us-west-2",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	cmd.Flags().String("region", "", "Cloud service provider region where the resources to be accessed using the private link attachment are located.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) privateLinkAttachmentCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}
	cloud = strings.ToUpper(cloud)

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	createPrivateLinkAttachment := networkingprivatelinkv1.NetworkingV1PrivateLinkAttachment{
		Spec: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentSpec{
			Environment: &networkingprivatelinkv1.ObjectReference{Id: environmentId},
			Cloud:       networkingprivatelinkv1.PtrString(cloud),
			Region:      networkingprivatelinkv1.PtrString(region),
		},
	}

	if name != "" {
		createPrivateLinkAttachment.Spec.SetDisplayName(name)
	}

	attachment, err := c.V2Client.CreatePrivateLinkAttachment(createPrivateLinkAttachment)
	if err != nil {
		return err
	}

	return printPrivateLinkAttachmentTable(cmd, attachment)
}
