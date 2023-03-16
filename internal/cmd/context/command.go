package context

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type command struct {
	*pcmd.CLICommand
	resolver pcmd.FlagResolver
}

func New(prerunner pcmd.PreRunner, resolver pcmd.FlagResolver) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "context",
		Aliases: []string{"ctx"},
		Short:   "Manage CLI configuration contexts.",
		Long:    "Manage CLI configuration contexts. Contexts define the state of a Confluent Cloud or Confluent Platform login.",
	}

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		resolver:   resolver,
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newUseCommand())

	return cmd
}

// context retrieves either a specific context or the current context.
func (c *command) context(args []string) (*dynamicconfig.DynamicContext, error) {
	if len(args) == 1 {
		return c.Config.FindContext(args[0])
	}

	if ctx := c.Config.Context(); ctx != nil {
		return ctx, nil
	} else {
		return nil, errors.NewErrorWithSuggestions("no context selected", "Select an existing context with `confluent context use`, or supply a specific context name as an argument.")
	}
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteContexts(c.Config.Config)
}
