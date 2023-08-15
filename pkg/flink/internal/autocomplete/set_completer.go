package autocomplete

import (
	prompt "github.com/confluentinc/go-prompt"
)

func SetCompleter(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "SET 'table.results-timeout' = '600';", Description: "Total amount of time in seconds to wait before timing out the request waiting for results to be ready."},
		{Text: "SET 'table.local-time-zone;' = 'Europe/Berlin';", Description: "Used to modify the configuration or list the configuration"},
	}

	return SuggestFromPrefix(s, in.TextBeforeCursor())
}
