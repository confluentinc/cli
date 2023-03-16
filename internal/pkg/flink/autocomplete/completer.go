package autocomplete

import (
	prompt "github.com/confluentinc/go-prompt"
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

func CompleterWithDocsExamples(getSmartCompletion func() bool) prompt.Completer {
	docsCompleter := generateDocsCompleter()
	return combineCompleters(getSmartCompletion, Completer, docsCompleter)
}

// Since we combine completers twice, we just need to control this properly once using "getSmartCompletion". Maybe
// we could solve this in a more elegant way, but this works for now.
func smartCompletionEnabled() bool {
	return true
}

func Completer(in prompt.Document) []prompt.Suggest {
	return combineCompleters(smartCompletionEnabled, examplesCompleter, setCompleter, showCompleter)(in)
}
