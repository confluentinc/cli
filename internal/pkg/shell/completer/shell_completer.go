package completer

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type ShellCompleter struct {
	*RootCompleter
	*SubCommandCompleter
}

func NewShellCompleter(rootCmd *cobra.Command, cliName string) *ShellCompleter {
	return &ShellCompleter{
		RootCompleter:        NewRootCompleter(rootCmd),
		SubCommandCompleter: NewSubCommandCompleter(rootCmd),
	}
}

func (c *ShellCompleter) Complete(d prompt.Document) []prompt.Suggest {
	return append(c.RootCompleter.Complete(d), c.SubCommandCompleter.Complete(d)...)
}
