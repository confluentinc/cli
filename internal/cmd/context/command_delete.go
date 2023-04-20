package context

import (
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
		Use:               "delete <context-1> [context-2] ... [context-n]",
		Short:             "Delete contexts.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

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
		if err := c.Config.DeleteContext(id); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.Context)

	return errs.ErrorOrNil()
}

func (c *command) confirmDeletion(cmd *cobra.Command, args []string) error {
	describeFunc := func(id string) error {
		_, err := c.Config.FindContext(id)
		return err
	}

	if err := deletion.ValidateArgsForDeletion(cmd, args, resource.Context, describeFunc); err != nil {
		return err
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.Context, args); err != nil || !ok {
		return err
	}

	return nil
}
