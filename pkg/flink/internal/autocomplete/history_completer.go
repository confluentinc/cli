package autocomplete

import (
	"github.com/confluentinc/go-prompt"
)

func GenerateHistoryCompleter(history []string) prompt.Completer {
	return func(in prompt.Document) []prompt.Suggest {
		historyCompletions := make([]prompt.Suggest, 0)
		// iterate backwards to show recent entries first
		for i := len(history) - 1; i >= 0; i-- {
			historyCompletions = append(historyCompletions, prompt.Suggest{Text: history[i], Description: "History entry"})
		}
		return SuggestNextWord(historyCompletions, in.TextBeforeCursor())
	}
}
