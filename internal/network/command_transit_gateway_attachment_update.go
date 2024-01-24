package network

import (
	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newTransitGatewayAttachmentUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing transit gateway attachment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validTransitGatewayAttachmentArgs),
		RunE:              c.transitGatewayAttachmentUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of transit gateway attachment "tgwa-123456".`,
				Code: `confluent network transit-gateway-attachment update tgwa-123456 --name "new name"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the transit gateway attachment.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) transitGatewayAttachmentUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updateTransitGatewayAttachment := networkingv1.NetworkingV1TransitGatewayAttachmentUpdate{
		Spec: &networkingv1.NetworkingV1TransitGatewayAttachmentSpecUpdate{
			DisplayName: networkingv1.PtrString(name),
			Environment: &networkingv1.ObjectReference{Id: environmentId},
		},
	}

	attachment, err := c.V2Client.UpdateTransitGatewayAttachment(environmentId, args[0], updateTransitGatewayAttachment)
	if err != nil {
		return err
	}

	return printTransitGatewayAttachmentTable(cmd, attachment)
}
