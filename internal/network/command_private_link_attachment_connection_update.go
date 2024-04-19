package network

import (
	"github.com/spf13/cobra"

	networkingprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-privatelink/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newPrivateLinkAttachmentConnectionUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing private link attachment connection.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.privateLinkAttachmentConnectionUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of private link attachment connection "plattc-123456".`,
				Code: `confluent network private-link attachment connection update plattc-123456 --name "new name"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the private link attachment connection.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("name"))

	return cmd
}

func (c *command) privateLinkAttachmentConnectionUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	updatePrivateLinkAttachmentConnection := networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionUpdate{
		Spec: &networkingprivatelinkv1.NetworkingV1PrivateLinkAttachmentConnectionSpecUpdate{
			DisplayName: networkingprivatelinkv1.PtrString(name),
			Environment: &networkingprivatelinkv1.ObjectReference{Id: environmentId},
		},
	}

	connection, err := c.V2Client.UpdatePrivateLinkAttachmentConnection(environmentId, args[0], updatePrivateLinkAttachmentConnection)
	if err != nil {
		return err
	}

	return printPrivateLinkAttachmentConnectionTable(cmd, connection)
}
