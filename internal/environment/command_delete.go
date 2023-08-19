package environment

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more Confluent Cloud environments.",
		Long:              "Delete one or more Confluent Cloud environments and all of their resources.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	environment, err := c.V2Client.GetOrgEnvironment(args[0])
	if err != nil {
		return resource.ResourcesNotFoundError(cmd, resource.Environment, args[0])
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetOrgEnvironment(id)
		return err == nil
	}

	if confirm, err := deletion.ValidateAndConfirmDeletionWithName(cmd, args, existenceFunc, resource.Environment, environment.GetDisplayName()); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteOrgEnvironment(id)
	}

	deletedIDs, err := deletion.Delete(args, deleteFunc, resource.Environment)

	errs := multierror.Append(err, c.deleteEnvironmentsFromConfig(deletedIDs))
	if errs.ErrorOrNil() != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(errors.ListResourceSuggestions, resource.Environment, "confluent environment"))
	}

	return nil
}

func (c *command) deleteEnvironmentsFromConfig(deletedIDs []string) error {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	for _, id := range deletedIDs {
		if id == c.Context.GetCurrentEnvironment() {
			c.Context.SetCurrentEnvironment("")
			errs = multierror.Append(errs, c.Config.Save())
		}
		c.Context.DeleteEnvironment(id)
		_ = c.Config.Save()
	}

	return errs.ErrorOrNil()
}
