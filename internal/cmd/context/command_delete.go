package context

import (
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
		Use:               "delete <context-1> [context-2] ... [context-N]",
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
			return err
		}
		return nil
	}

	// Multiple
	contextSet := types.NewSet()
	for _, context := range c.Config.Contexts {
		contextSet.Add(context.Name)
	}

	invalidContexts := contextSet.Difference(args)
	if len(invalidContexts) > 0 {
		return errors.New("context(s) not found: " + utils.ArrayToCommaDelimitedStringWithAnd(invalidContexts))
	}

	return nil
}
