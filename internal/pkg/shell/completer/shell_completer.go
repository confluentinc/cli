package completer

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type ShellCompleter struct {
	*CobraCompleter
}

func NewShellCompleter(rootCmd *cobra.Command, cliName string) *ShellCompleter {
	return &ShellCompleter{
		CobraCompleter:      NewCobraCompleter(rootCmd),
	}
}

func (c *ShellCompleter) Complete(d prompt.Document) []prompt.Suggest {
	return c.CobraCompleter.Complete(d)
}
