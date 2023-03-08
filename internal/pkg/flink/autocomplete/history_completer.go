package autocomplete

import (
	"github.com/c-bata/go-prompt"
)

// Currently disabled. History entries are not shown in autocompletion.
//  If we enable this again in the future, it would be good to filter duplicates in the history.
func generateHISTORYCompleter(history []string) prompt.Completer {
	historyCompletions := []prompt.Suggest{}
	for _, v := range history {
		historyCompletions = append(historyCompletions, prompt.Suggest{Text: v, Description: "History entry"})
	}

	return func(in prompt.Document) []prompt.Suggest {
		return SuggestFromPrefix(historyCompletions, in.TextBeforeCursor())
	}
}
