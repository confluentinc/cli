package completer

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type ShellCompleter struct {
	*CobraCompleter
	ServerSideCompleter
}

func NewShellCompleter(rootCmd *cobra.Command) *ShellCompleter {
	return &ShellCompleter{
		CobraCompleter:      NewCobraCompleter(rootCmd),
		ServerSideCompleter: NewServerSideCompleter(rootCmd),
	}
}

func (c *ShellCompleter) Complete(d prompt.Document) []prompt.Suggest {
	return append(c.CobraCompleter.Complete(d), c.ServerSideCompleter.Complete(d)...)
}
