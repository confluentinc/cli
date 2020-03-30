package completer

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type MasterCompleter struct {
	*FuzzyCompleter
	*CoreCompleter
	*BackgroundCommandCompleter
	FuzzyComplete bool
}

func NewMasterCompleter(rootCmd *cobra.Command, cliName string) *MasterCompleter {
	return &MasterCompleter{
		FuzzyCompleter:             NewFuzzyCompleter(rootCmd, cliName),
		CoreCompleter:              NewCoreCompleter(rootCmd),
		BackgroundCommandCompleter: NewBackgroundCommandCompleter(rootCmd, cliName, DefaultUpdateInterval),
		FuzzyComplete:              false,
	}
}

func (c *MasterCompleter) Complete(d prompt.Document) []prompt.Suggest {
	coreCompletions := c.CoreCompleter.Complete(d)
	if c.FuzzyComplete {
		return c.FuzzyCompleter.Complete(d)
		return prompt.FilterFuzzy(coreCompletions, d.GetWordBeforeCursor(), true)
	}
	return append(coreCompletions, c.BackgroundCommandCompleter.Complete(d)...)
}

func (c *MasterCompleter) SetRootCmd(rootCmd *cobra.Command) {
	c.FuzzyCompleter.RootCmd = rootCmd
	c.CoreCompleter.RootCmd = rootCmd
	c.CoreCommandCompleter.RootCmd = rootCmd
}
