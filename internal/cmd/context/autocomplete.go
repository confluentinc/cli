package context

import (
	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return autocompleteContexts(c.Config.Config)
}

func autocompleteContexts(cfg *v1.Config) []string {
	suggestions := make([]string, len(cfg.Contexts))
	i := 0
	for context := range cfg.Contexts {
		suggestions[i] = context
		i++
	}
	return suggestions
}
