package context

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
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

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	if err := c.checkExistence(cmd, args); err != nil {
		return err
	}

	if ok, err := form.ConfirmDeletionYesNo(cmd, resource.Context, args); err != nil || !ok {
		return err
	}

	var errs error
	for _, ctxName := range args {
		if err := c.Config.DeleteContext(ctxName); err != nil {
			errs = errors.Join(errs, err)
		} else {
			output.Printf(errors.DeletedResourceMsg, resource.Context, ctxName)
		}
	}

	return errs
}

func (c *command) checkExistence(cmd *cobra.Command, args []string) error {
	// Single
	if len(args) == 1 {
		if _, err := c.Config.FindContext(args[0]); err != nil {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Context, args[0]), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Context))
		}
		return nil
	}

	// Multiple
	set := types.NewSet()
	for _, context := range c.Config.Contexts {
		set.Add(context.Name)
	}

	validArgs, invalidArgs := set.IntersectionAndDifference(args)
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return err
	} else if force && len(invalidArgs) > 0 {
		args = validArgs
		return nil
	}

	invalidArgsStr := utils.ArrayToCommaDelimitedString(invalidArgs, "and")
	if len(invalidArgs) == 1 {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Context, invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Context))
	} else if len(invalidArgs) > 1 {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.NotFoundErrorMsg, resource.Plural(resource.Context), invalidArgsStr), fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Context))
	}

	return nil
}
