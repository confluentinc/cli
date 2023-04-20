package environment

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete Confluent Cloud environments and all their resources.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	if err := c.confirmDeletion(cmd, args); err != nil {
		return err
	}

	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deleted []string
	for _, id := range args {
		if err := c.V2Client.DeleteOrgEnvironment(id); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deleted = append(deleted, id)
			if err := c.deletePostProcess(id); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}
	deletion.PrintSuccessMsg(deleted, resource.Environment)

	if errs.ErrorOrNil() != nil {
		return errors.NewErrorWithSuggestions(errs.Error(), fmt.Sprintf(errors.ListResourceSuggestions, resource.Environment, pcmd.FullParentName(cmd)))
	}

	return nil
}

func (c *command) confirmDeletion(cmd *cobra.Command, args []string) error {
	var displayName string
	describeFunc := func(id string) error {
		environment, err := c.V2Client.GetOrgEnvironment(id)
		if err == nil && id == args[0] {
			displayName = environment.GetDisplayName()
		}
		return err
	}

	if err := deletion.ValidateArgs(cmd, args, resource.Environment, describeFunc); err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.Environment, args[0], displayName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.Environment, args); err != nil || !ok {
			return err
		}
	}

	return nil
}

func (c *command) deletePostProcess(id string) error {
	var err error
	if id == c.Context.GetCurrentEnvironment() {
		c.Context.SetCurrentEnvironment("")

		if err2 := c.Config.Save(); err2 != nil {
			err = errors.Wrap(err2, errors.EnvSwitchErrorMsg)
		}
	}
	c.Context.DeleteEnvironment(id)
	_ = c.Config.Save()

	return err
}
