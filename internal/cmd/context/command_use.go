package context

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *command) newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "use <name>",
		Short:             "Use a context.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	name := args[0]

	if err := c.Config.UseContext(name); err != nil {
		return err
	}

	cmd.Printf("Using context \"%s\".\n", name)
	return nil
}
