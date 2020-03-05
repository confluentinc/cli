package completer

import (
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type CoreCommandCompleter struct {
	RootCmd *cobra.Command

	// map[string]CompletionFunc
	completionFuncsByCmdName *sync.Map
}

func NewCoreCommandCompleter(rootCmd *cobra.Command) *CoreCommandCompleter {
	return &CoreCommandCompleter{
		RootCmd:                  rootCmd,
		completionFuncsByCmdName: new(sync.Map),
	}
}

func (c *CoreCommandCompleter) Complete(d prompt.Document) []prompt.Suggest {
	cmd := c.RootCmd
	args := strings.Fields(d.CurrentLine())

	if found, _, err := cmd.Find(args); err == nil {
		cmd = found
	}
	f, ok := c.completionFuncsByCmdName.Load(commandKey(cmd))
	if !ok {
		return []prompt.Suggest{}
	}
	completionFunc := f.(CompletionFunc)
	return completionFunc(d)
}

func (c *CoreCommandCompleter) AddCommand(cmd *cobra.Command, completionFunc CompletionFunc) {
	c.completionFuncsByCmdName.Store(commandKey(cmd), completionFunc)
}

func commandKey(cmd *cobra.Command) string {
	return cmd.CommandPath()
}
