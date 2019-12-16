package cmd

import (
	"sync"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

const (
	maxCacheAge = 10 * time.Second
)

type Completer struct {
	SuggestionFunctionsByCommand *sync.Map //map[string]func() []prompt.Suggest
	cachedSuggestionsByCommand   *sync.Map //map[string][]prompt.Suggest
	lastFetchTimeByCommand       *sync.Map //map[string]time.Time
}

func NewCompleter() *Completer {
	c := &Completer{
		SuggestionFunctionsByCommand: new(sync.Map),
		cachedSuggestionsByCommand:   new(sync.Map),
		lastFetchTimeByCommand:       new(sync.Map),
	}
	go c.runSuggestions()
	return c
}

func (c *Completer) Complete(annotation string, d prompt.Document) []prompt.Suggest {
	cachedSuggestions, ok := c.cachedSuggestionsByCommand.Load(annotation)
	if !ok {
		return []prompt.Suggest{}
	}
	if suggestions, ok := cachedSuggestions.([]prompt.Suggest); ok {
		return suggestions
	} else {
		return []prompt.Suggest{}
	}
}

func (c *Completer) AddSuggestionFunction(cmd *cobra.Command, sFunc func() []prompt.Suggest) {
	annotation := cmd.Annotations[CALLBACK_ANNOTATION]
	c.SuggestionFunctionsByCommand.Store(annotation, sFunc)
}

func (c *Completer) runSuggestions() {
	for range time.Tick(maxCacheAge) {
		c.UpdateAllSuggestions()
	}
}

func (c *Completer) UpdateAllSuggestions() {
	c.SuggestionFunctionsByCommand.Range(func(key, sFunc interface{}) bool {
		annotation := key.(string)                        // Will panic if not string.
		suggestionFunc := sFunc.(func() []prompt.Suggest) // Will also cause panic if not the correct type.
		c.updateSuggestion(annotation, suggestionFunc)
		return true
	})
}

func (c *Completer) updateSuggestion(annotation string, suggestionFunc func() []prompt.Suggest) {
	c.lastFetchTimeByCommand.Store(annotation, time.Now())
	suggestions := suggestionFunc()
	c.cachedSuggestionsByCommand.Store(annotation, suggestions)
}
