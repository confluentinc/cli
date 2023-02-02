package components

import (
	prompt "github.com/c-bata/go-prompt"
)

func CombineCompleters(completers ...prompt.Completer) prompt.Completer {
	return func(d prompt.Document) []prompt.Suggest {
		var suggestions []prompt.Suggest
		for _, c := range completers {
			suggestions = append(suggestions, c(d)...)
		}
		return suggestions
	}
}

// Not currently used in favor of completerWithHistory, but could be used to provide a
// completer without history if we want to give the user the possibility of disabling that
func completer(in prompt.Document) []prompt.Suggest {
	return CombineCompleters(EXAMPLESCompleter, SETCompleter, SHOWCompleter)(in)
}
