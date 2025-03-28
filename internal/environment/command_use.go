package environment

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/plural"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Use an environment in subsequent commands.",
		Long:              "Choose a Confluent Cloud environment to be used in subsequent commands which support passing an environment with the `--environment` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
	}

	return cmd
}

func (c *command) use(_ *cobra.Command, args []string) error {
	id := args[0]

	if _, err := c.V2Client.GetOrgEnvironment(id); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(errors.ListResourceSuggestions, plural.Plural(resource.Environment), "confluent environment"))
	}

	c.Context.SetCurrentEnvironment(id)
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UsingResourceMsg, resource.Environment, id)
	return nil
}
