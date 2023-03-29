package context

import (
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
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddSkipInvalidFlag(cmd)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	if validArgs, err := c.validateArgs(cmd, args); err != nil {
		return err
	} else {
		args = validArgs
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.Context, args); err != nil || !ok {
		return err
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if err := c.Config.DeleteContext(id); err != nil {
			errs = errors.Join(errs, err)
		} else {
			deleted = append(deleted, id)
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.Context)

	return errs
}

func (c *command) validateArgs(cmd *cobra.Command, args []string) ([]string, error) {
	describeFunc := func(id string) error {
		if _, err := c.Config.FindContext(id); err != nil {
			return err
		}
		return nil
	}

	return deletion.ValidateArgsForDeletion(cmd, args, resource.Context, describeFunc)
}
