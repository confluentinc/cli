package environment

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
	if confirm, err := c.confirmDeletion(cmd, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteOrgEnvironment(id)
	}

	deletedIDs, err := resource.Delete(args, deleteFunc, resource.Environment)
	if err2 := c.deleteEnvironmentsFromConfig(deletedIDs); err2 != nil {
		err = multierror.Append(err, err2)
	}
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(errors.ListResourceSuggestions, resource.Environment, pcmd.FullParentName(cmd)))
	}

	return nil
}

func (c *command) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	var displayName string
	describeFunc := func(id string) error {
		environment, err := c.V2Client.GetOrgEnvironment(id)
		if err != nil {
			return err
		}
		if id == args[0] {
			displayName = environment.GetDisplayName()
		}

		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.Environment, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.Environment, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.Environment, args[0], displayName), displayName); err != nil {
		return false, err
	}

	return true, nil
}

func (c *command) deleteEnvironmentsFromConfig(deletedIDs []string) error {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	for _, id := range deletedIDs {
		if id == c.Context.GetCurrentEnvironment() {
			c.Context.SetCurrentEnvironment("")
			if err := c.Config.Save(); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
		c.Context.DeleteEnvironment(id)
		_ = c.Config.Save()
	}

	return errs.ErrorOrNil()
}
