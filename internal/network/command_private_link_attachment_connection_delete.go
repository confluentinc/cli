package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/plural"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *command) newPrivateLinkAttachmentConnectionDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <id-1> [id-2] ... [id-n]",
		Short: "Delete one or more private link attachment connections.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.privateLinkAttachmentConnectionDelete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) privateLinkAttachmentConnectionDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetPrivateLinkAttachmentConnection(environmentId, id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.PrivateLinkAttachmentConnection); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeletePrivateLinkAttachmentConnection(environmentId, id)
	}

	deletedIds, err := deletion.DeleteWithoutMessage(cmd, args, deleteFunc)
	deleteMsg := "Requested to delete %s %s.\n"
	if len(deletedIds) == 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, resource.PrivateLinkAttachmentConnection, fmt.Sprintf(`"%s"`, deletedIds[0]))
	} else if len(deletedIds) > 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, plural.Plural(resource.PrivateLinkAttachmentConnection), utils.ArrayToCommaDelimitedString(deletedIds, "and"))
	}

	return err
}
