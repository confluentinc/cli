package context

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <context-1> [context-2] ... [context-n]",
		Short:             "Delete one or more contexts.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
	}

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
		return c.Config.DeleteContext(id)
	}

	_, err := resource.Delete(args, deleteFunc, resource.Context)
	return err
}

func (c *command) confirmDeletion(cmd *cobra.Command, args []string) (bool, error) {
	describeFunc := func(id string) error {
		_, err := c.Config.FindContext(id)
		return err
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.Context, describeFunc); err != nil {
		return false, err
	}

	return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.Context, args))
}
