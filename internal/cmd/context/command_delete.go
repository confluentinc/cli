package context

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <context>",
		Short:             "Delete a context.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}
	cmd.Flags().Bool("force", false, "Skip the deletion confirmation prompt.")

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	ctx, err := c.Config.FindContext(args[0])
	if err != nil {
		return err
	}

	if confirm, err := form.ConfirmDeletionYesNo(cmd, resource.Context, ctx.Name); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	if err := c.Config.DeleteContext(ctx.Name); err != nil {
		return err
	}

	utils.Printf(cmd, errors.DeletedResourceMsg, resource.Context, ctx.Name)
	return nil
}
