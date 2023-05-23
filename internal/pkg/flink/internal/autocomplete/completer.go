package autocomplete

import (
	"github.com/confluentinc/go-prompt"
)

func combineCompleters(getSmartCompletion func() bool, completers ...prompt.Completer) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		suggestions := []prompt.Suggest{}

		if !getSmartCompletion() {
			return suggestions
		}

		for _, c := range completers {
			suggestions = append(suggestions, c(d)...)
		}
		return suggestions
	}
}

type completerBuilder struct {
	isSmartCompletionEnabled func() bool
	completer                prompt.Completer
}

func NewCompleterBuilder(isSmartCompletionEnabled func() bool) *completerBuilder {
	return &completerBuilder{
		isSmartCompletionEnabled: isSmartCompletionEnabled,
	}
}

func (c *completerBuilder) AddCompleter(completer prompt.Completer) *completerBuilder {
	if c.completer == nil {
		c.completer = combineCompleters(c.isSmartCompletionEnabled, completer)
	} else {
		c.completer = combineCompleters(c.isSmartCompletionEnabled, c.completer, completer)
	}
	return c
}

func (a *completerBuilder) BuildCompleter() prompt.Completer {
	return a.completer
}
