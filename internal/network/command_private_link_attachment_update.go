package network

import (
	"github.com/spf13/cobra"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPrivateLinkAttachmentUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing private link attachment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPrivateLinkAttachmentArgs),
		RunE:              c.privateLinkAttachmentUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of private link attachment "platt-123456".`,
				Code: `confluent network private-link attachment update platt-123456 --name "new name"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the private link attachment.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) privateLinkAttachmentUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updatePrivateLinkAttachment := networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentUpdate{
		Spec: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentSpecUpdate{
			DisplayName: networkingprivatelinkv1.PtrString(name),
			Environment: &networkingprivatelinkv1.ObjectReference{Id: environmentId},
		},
	}

	attachment, err := c.V2Client.UpdatePrivateLinkAttachment(environmentId, args[0], updatePrivateLinkAttachment)
	if err != nil {
		return err
	}

	return printPrivateLinkAttachmentTable(cmd, attachment)
}
