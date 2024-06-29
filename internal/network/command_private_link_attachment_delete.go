package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func (c *command) newPrivateLinkAttachmentDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more private link attachments.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPrivateLinkAttachmentArgsMultiple),
		RunE:              c.privateLinkAttachmentDelete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) privateLinkAttachmentDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetPrivateLinkAttachment(environmentId, id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.PrivateLinkAttachment); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeletePrivateLinkAttachment(environmentId, id); err != nil {
			return fmt.Errorf(errors.DeleteResourceErrorMsg, resource.PrivateLinkAttachment, id, err)
		}
		return nil
	}

	deletedIds, err := deletion.DeleteWithoutMessage(args, deleteFunc)
	deleteMsg := "Requested to delete %s %s.\n"
	if len(deletedIds) == 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, resource.PrivateLinkAttachment, fmt.Sprintf(`"%s"`, deletedIds[0]))
	} else if len(deletedIds) > 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, resource.Plural(resource.PrivateLinkAttachment), utils.ArrayToCommaDelimitedString(deletedIds, "and"))
	}

	return err
}
