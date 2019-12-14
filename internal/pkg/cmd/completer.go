package cmd

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type Completer struct {
	SuggestionFunctionsByCommand map[string]func() []prompt.Suggest
}

func NewCompleter() *Completer {
	return &Completer{SuggestionFunctionsByCommand: map[string]func() []prompt.Suggest{}}
}

func (c *Completer) Complete(annotation string, d prompt.Document) []prompt.Suggest {
	sFunc := c.SuggestionFunctionsByCommand[annotation]
	if sFunc != nil {
		return sFunc()
	}
	return nil
}

func (c *Completer) AddSuggestionFunction(cmd *cobra.Command, sFunc func() []prompt.Suggest) {
	key := cmd.Annotations[CALLBACK_ANNOTATION]
	c.SuggestionFunctionsByCommand[key] = sFunc
}
