package completer

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type MasterCompleter struct {
	*FuzzyCompleter
	*CoreCompleter
	*CoreCommandCompleter
	FuzzyComplete bool
}

func NewMasterCompleter(rootCmd *cobra.Command, cliName string) *MasterCompleter {
	return &MasterCompleter{
		FuzzyCompleter:       NewFuzzyCompleter(rootCmd, cliName),
		CoreCompleter:        NewCoreCompleter(rootCmd),
		CoreCommandCompleter: NewCoreCommandCompleter(rootCmd),
		FuzzyComplete:        false,
	}
}

func (c *MasterCompleter) Complete(d prompt.Document) []prompt.Suggest {
	if c.FuzzyComplete {
		return c.FuzzyCompleter.Complete(d)
	}
	return append(c.CoreCompleter.Complete(d), c.CoreCommandCompleter.Complete(d)...)
}
