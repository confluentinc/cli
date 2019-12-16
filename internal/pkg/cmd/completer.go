package cmd

import (
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

const (
	maxCacheAge = 3 * time.Second
)

type Completer struct {
	SuggestionFunctionsByCommand map[string]func() []prompt.Suggest
	cachedSuggestionsByCommand   map[string][]prompt.Suggest
	lastFetchTimeByCommand       map[string]time.Time
}

func NewCompleter() *Completer {
	c := &Completer{
		SuggestionFunctionsByCommand: map[string]func() []prompt.Suggest{},
		cachedSuggestionsByCommand:   map[string][]prompt.Suggest{},
		lastFetchTimeByCommand:       map[string]time.Time{},
	}
	go c.runSuggestions()
	return c
}

func (c *Completer) Complete(annotation string, d prompt.Document) []prompt.Suggest {
	sFunc := c.SuggestionFunctionsByCommand[annotation]
	if sFunc != nil {
		return c.cachedSuggestionsByCommand[annotation]
	}
	return nil
}

func (c *Completer) AddSuggestionFunction(cmd *cobra.Command, sFunc func() []prompt.Suggest) {
	key := cmd.Annotations[CALLBACK_ANNOTATION]
	c.SuggestionFunctionsByCommand[key] = sFunc
}

func (c *Completer) runSuggestions() {
	for range time.Tick(maxCacheAge) {
		c.updateAllSuggestions()
	}
}

func (c *Completer) updateAllSuggestions() {
	for annotation, sFunc := range c.SuggestionFunctionsByCommand {
		c.lastFetchTimeByCommand[annotation] = time.Now()
		c.cachedSuggestionsByCommand[annotation] = sFunc()
	}
}
