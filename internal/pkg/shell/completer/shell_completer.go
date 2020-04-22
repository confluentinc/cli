package completer

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type ShellCompleter struct {
	*CobraCompleter
	*SubCommandCompleter
}

func NewShellCompleter(rootCmd *cobra.Command, cliName string) *ShellCompleter {
	return &ShellCompleter{
		CobraCompleter:      NewCobraCompleter(rootCmd),
		SubCommandCompleter: NewSubCommandCompleter(rootCmd),
	}
}

func (c *ShellCompleter) Complete(d prompt.Document) []prompt.Suggest {
	return append(c.CobraCompleter.Complete(d), c.SubCommandCompleter.Complete(d)...)
}
