package autocomplete

import (
	"github.com/c-bata/go-prompt"
)

func generateHISTORYCompleter(history []string) prompt.Completer {
	historyCompletions := []prompt.Suggest{}
	for _, v := range history {
		historyCompletions = append(historyCompletions, prompt.Suggest{Text: v, Description: "History entry"})
	}

	return func(in prompt.Document) []prompt.Suggest {
		return prompt.FilterHasPrefix(historyCompletions, in.TextBeforeCursor(), true)
	}
}
