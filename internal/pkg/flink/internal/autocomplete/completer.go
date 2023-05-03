package autocomplete

import (
	"github.com/confluentinc/go-prompt"
)

func combineCompleters(getSmartCompletion func() bool, completers ...prompt.Completer) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		suggestions := []prompt.Suggest{}

		if !(getSmartCompletion()) {
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

func (a *completerBuilder) AddCompleter(completer prompt.Completer) *completerBuilder {
	if a.completer == nil {
		a.completer = combineCompleters(a.isSmartCompletionEnabled, completer)
	} else {
		a.completer = combineCompleters(a.isSmartCompletionEnabled, a.completer, completer)
	}
	return a
}

func (a *completerBuilder) BuildCompleter() prompt.Completer {
	return a.completer
}
