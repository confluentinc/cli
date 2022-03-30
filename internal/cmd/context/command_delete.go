package context

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "delete <context>",
		Short:             "Delete a context.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.delete),
	}
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	ctx, err := c.Config.FindContext(args[0])
	if err != nil {
		return err
	}

	if err := c.Config.DeleteContext(ctx.Name); err != nil {
		return err
	}

	utils.Printf(cmd, "Deleted context \"%s\".\n", ctx.Name)
	return nil
}
