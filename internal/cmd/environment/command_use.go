package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Switch to the specified Confluent Cloud environment.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	id := args[0]

	if _, err := c.V2Client.GetOrgEnvironment(id); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), "List available environments with `confluent environment list`.")
	}

	c.Context.SetCurrentEnvironment(id)
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf("Now using \"%s\" as the default (active) environment.\n", id)
	return nil
}
