package environment

import (
	"context"

	"github.com/c-bata/go-prompt"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteEnvironments(c.Client)
}

func (c *command) Cmd() *cobra.Command {
	return c.Command
}

func (c *command) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}

func (c *command) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	environments, err := c.Client.Account.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return suggestions
	}
	for _, env := range environments {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        env.Id,
			Description: env.Name,
		})
	}
	return suggestions
}
