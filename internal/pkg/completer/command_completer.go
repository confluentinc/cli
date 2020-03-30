package completer

import (
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type CoreCommandCompleter struct {
	RootCmd *cobra.Command
	cliName string

	// map[string]CommandCompletionFunc
	completionFuncsByCmdName *sync.Map
}

func NewCoreCommandCompleter(rootCmd *cobra.Command, cliName string) *CoreCommandCompleter {
	return &CoreCommandCompleter{
		RootCmd:                  rootCmd,
		cliName:                  cliName,
		completionFuncsByCmdName: new(sync.Map),
	}
}

func (c *CoreCommandCompleter) Complete(d prompt.Document) []prompt.Suggest {
	cmd := c.findCommand(d)
	key := c.commandKey(cmd)
	f, ok := c.completionFuncsByCmdName.Load(key)
	if !ok {
		return []prompt.Suggest{}
	}
	cmdCompletionFunc := f.(CommandCompletionFunc)
	return prompt.FilterHasPrefix(cmdCompletionFunc(), d.GetWordBeforeCursor(), true)
}

func (c *CoreCommandCompleter) AddCommand(cmd *cobra.Command, cmdCompletionFunc CommandCompletionFunc) {
	c.completionFuncsByCmdName.Store(c.commandKey(cmd), cmdCompletionFunc)
}

func (c *CoreCommandCompleter) findCommand(d prompt.Document) *cobra.Command {
	cmd := c.RootCmd
	args := strings.Fields(d.CurrentLine())

	if found, _, err := cmd.Find(args); err == nil {
		cmd = found
	}
	return cmd
}

func (c *CoreCommandCompleter) commandKey(cmd *cobra.Command) string {
	key := strings.TrimPrefix(cmd.CommandPath(), c.cliName)
	key = strings.TrimSpace(key)
	return key
}
