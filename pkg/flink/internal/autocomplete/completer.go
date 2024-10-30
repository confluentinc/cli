package autocomplete

import (
	"github.com/confluentinc/go-prompt"
)

func combineCompleters(getCompletionsEnabled func() bool, completers ...prompt.Completer) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		suggestions := []prompt.Suggest{}

		if !getCompletionsEnabled() {
			return suggestions
		}

		for _, c := range completers {
			suggestions = append(suggestions, c(d)...)
		}
		return suggestions
	}
}

type completerBuilder struct {
	completionsEnabled func() bool
	completer          prompt.Completer
}

func NewCompleterBuilder(completionsEnabled func() bool) *completerBuilder {
	return &completerBuilder{completionsEnabled: completionsEnabled}
}

func (c *completerBuilder) AddCompleter(completer prompt.Completer) *completerBuilder {
	if c.completer == nil {
		c.completer = combineCompleters(c.completionsEnabled, completer)
	} else {
		c.completer = combineCompleters(c.completionsEnabled, c.completer, completer)
	}
	return c
}

func (c *completerBuilder) BuildCompleter() prompt.Completer {
	return c.completer
}
