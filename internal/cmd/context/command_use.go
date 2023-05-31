package context

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *command) newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "use <context>",
		Short:             "Choose a context to be used in subsequent commands.",
		Long:              "Choose a context to be used in subsequent commands which support passing a context with the `--context` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	if err := c.Config.UseContext(args[0]); err != nil {
		return err
	}

	cmd.Printf("Using context \"%s\".\n", args[0])
	return nil
}
