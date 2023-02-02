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

func completerWithHistory(history []string) prompt.Completer {
	HISTORYCompleter := generateHISTORYCompleter(history)
	return CombineCompleters(completer, HISTORYCompleter)
}

func completer(in prompt.Document) []prompt.Suggest {
	return CombineCompleters(EXAMPLESCompleter, SETCompleter, SHOWCompleter)(in)
}
