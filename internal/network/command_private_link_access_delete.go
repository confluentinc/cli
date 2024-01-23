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

func (c *command) newPrivateLinkAccessDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more private link accesses.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validPrivateLinkAccessArgsMultiple),
		RunE:              c.privateLinkAccessDelete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) privateLinkAccessDelete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetPrivateLinkAccess(environmentId, id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.PrivateLinkAccess); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeletePrivateLinkAccess(environmentId, id); err != nil {
			return fmt.Errorf(errors.DeleteResourceErrorMsg, resource.PrivateLinkAccess, id, err)
		}
		return nil
	}

	deletedIds, err := deletion.DeleteWithoutMessage(args, deleteFunc)
	deleteMsg := "Requested to delete %s %s.\n"
	if len(deletedIds) == 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, resource.PrivateLinkAccess, fmt.Sprintf(`"%s"`, deletedIds[0]))
	} else if len(deletedIds) > 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, resource.Plural(resource.PrivateLinkAccess), utils.ArrayToCommaDelimitedString(deletedIds, "and"))
	}

	return err
}
