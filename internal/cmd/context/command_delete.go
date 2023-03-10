package context

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	perrors "github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/set"
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

	promptMsg := fmt.Sprintf(perrors.DeleteResourceConfirmYesNoMsg, resource.Context, args[0])
	if len(args) > 1 {
		promptMsg = fmt.Sprintf(perrors.DeleteResourcesConfirmYesNoMsg, resource.Context, utils.ArrayToCommaDelimitedStringWithAnd(args))
	}
	if ok, err := form.ConfirmDeletion(cmd, promptMsg, ""); err != nil || !ok {
		return err
	}

	var errs error
	for _, ctxName := range args {
		if err := c.Config.DeleteContext(ctxName); err != nil {
			return err
		} else {
			output.Printf(perrors.DeletedResourceMsg, resource.Context, ctxName)
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
	contextSet := set.New()
	for _, context := range c.Config.Contexts {
		contextSet.Add(context.Name)
	}

	invalidContexts := contextSet.Difference(args)
	if len(invalidContexts) > 0 {
		return perrors.New("context(s) not found: " + utils.ArrayToCommaDelimitedStringWithAnd(invalidContexts))
	}

	return nil
}
