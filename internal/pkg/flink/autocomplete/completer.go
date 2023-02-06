package autocomplete

import (
	prompt "github.com/c-bata/go-prompt"
)

func combineCompleters(completers ...prompt.Completer) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		var suggestions []prompt.Suggest
		for _, c := range completers {
			suggestions = append(suggestions, c(d)...)
		}
		return suggestions
	}
}

func CompleterWithHistory(history []string) prompt.Completer {
	HISTORYCompleter := generateHISTORYCompleter(history)
	return combineCompleters(Completer, HISTORYCompleter)
}

func Completer(in prompt.Document) []prompt.Suggest {
	return combineCompleters(examplesCompleter, setCompleter, showCompleter)(in)
}
