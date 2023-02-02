package components

import (
	prompt "github.com/c-bata/go-prompt"
)

type CompleterFunc func(prompt.Document) []prompt.Suggest

func CombineCompleters(completers ...CompleterFunc) CompleterFunc {
	return func(d prompt.Document) []prompt.Suggest {
		var suggestions []prompt.Suggest
		for _, c := range completers {
			suggestions = append(suggestions, c(d)...)
		}
		return suggestions
	}
}

func completer(in prompt.Document) []prompt.Suggest {
	return CombineCompleters(EXAMPLESCompleter, SETCompleter, SHOWCompleter)(in)
}
