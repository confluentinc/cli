package environment

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	presource "github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id|name>",
		Short:             "Use an environment in subsequent commands.",
		Long:              "Choose a Confluent Cloud environment to be used in subsequent commands which support passing an environment with the `--environment` flag.",
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
		id, err = presource.ConvertEnvironmentNameToId(id, c.V2Client)
		if err != nil {
			return err
		}
		if _, err = c.V2Client.GetOrgEnvironment(id); err != nil {
			return errors.NewErrorWithSuggestions(err.Error(), "List available environments with `confluent environment list`.")
		}
	}

	c.Context.SetCurrentEnvironment(id)
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(errors.UsingResourceMsg, presource.Environment, id)
	return nil
}
