package completer

import (
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type SubCommandCompleter struct {
	RootCmd *cobra.Command

	// map[string]CompletionFunc
	completionFuncsByCmdName *sync.Map
}

func NewSubCommandCompleter(rootCmd *cobra.Command) *SubCommandCompleter {
	return &SubCommandCompleter{
		RootCmd:                  rootCmd,
		completionFuncsByCmdName: new(sync.Map),
	}
}

func (c *SubCommandCompleter) Complete(d prompt.Document) []prompt.Suggest {
	cmd := c.RootCmd
	args := strings.Fields(d.CurrentLine())

	if found, _, err := cmd.Find(args); err == nil {
		cmd = found
	}

	f, ok := c.completionFuncsByCmdName.Load(cmd.CommandPath())
	if !ok {
		return []prompt.Suggest{}
	}
	completionFunc := f.(CompletionFunc)
	return completionFunc(d)
}

func (c *SubCommandCompleter) AddCommand(cmd *cobra.Command, completionFunc CompletionFunc) {
	c.completionFuncsByCmdName.Store(cmd.CommandPath(), completionFunc)
}
