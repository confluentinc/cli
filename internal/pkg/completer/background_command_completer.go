package completer

import (
	"sync"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

const (
	DefaultUpdateInterval = 10 * time.Second
)

type BackgroundCommandCompleter struct {
	*CoreCommandCompleter
	updateInterval time.Duration

	//map[string][]prompt.Suggest
	cachedSuggestionsByCommand *sync.Map
}

func NewBackgroundCommandCompleter(rootCmd *cobra.Command, cliName string,
	updateInterval time.Duration) *BackgroundCommandCompleter {
	bc := &BackgroundCommandCompleter{
		CoreCommandCompleter:       NewCoreCommandCompleter(rootCmd, cliName),
		updateInterval:             updateInterval,
		cachedSuggestionsByCommand: new(sync.Map),
	}
	go bc.runSuggestions()
	return bc
}

func (c *BackgroundCommandCompleter) Complete(d prompt.Document) []prompt.Suggest {
	cmd := c.findCommand(d)
	key := c.commandKey(cmd)
	suggestions, ok := c.cachedSuggestionsByCommand.Load(key)
	if !ok {
		suggestions = c.CoreCommandCompleter.Complete(d)
		c.cachedSuggestionsByCommand.Store(key, suggestions)
	}
	if suggestions, ok := suggestions.([]prompt.Suggest); ok {
		return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
	} else {
		return []prompt.Suggest{}
	}
}

func (c *BackgroundCommandCompleter) AddSuggestionFunction(cmd *cobra.Command, cmdCompletionFunc CommandCompletionFunc) {
	c.CoreCommandCompleter.AddCommand(cmd, cmdCompletionFunc)
	key := c.commandKey(cmd)
	suggestions := cmdCompletionFunc()
	c.cachedSuggestionsByCommand.Store(key, suggestions)
}

func (c *BackgroundCommandCompleter) runSuggestions() {
	for range time.Tick(c.updateInterval) {
		c.UpdateAllSuggestions()
	}
}

func (c *BackgroundCommandCompleter) UpdateAllSuggestions() {
	c.completionFuncsByCmdName.Range(func(key, sFunc interface{}) bool {
		cmdKey := key.(string)                             // Will panic if not string.
		cmdCompletionFunc := sFunc.(CommandCompletionFunc) // Will also cause panic if not the correct type.
		suggestions := cmdCompletionFunc()
		c.cachedSuggestionsByCommand.Store(cmdKey, suggestions)
		return true
	})
}
