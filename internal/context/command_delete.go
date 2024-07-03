package context

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
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
	existenceFunc := func(id string) bool {
		_, err := c.Config.FindContext(id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.Context); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.Config.DeleteContext(id)
	}

	_, err := deletion.Delete(args, deleteFunc, resource.Context)
	return err
}
