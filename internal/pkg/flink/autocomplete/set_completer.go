package autocomplete

import (
	prompt "github.com/c-bata/go-prompt"
)

func setCompleter(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "SET 'pipeline.name' = 'SqlJob';", Description: "Sets the job name"},
		{Text: "SET 'parallelism.default' = '100';", Description: "Sets the job parallelism"},
		{Text: "SET 'sql-client.execution.result-mode' = 'TABLE';", Description: "Determines how the query result should be displayed"},
	}

	return prompt.FilterHasPrefix(s, in.TextBeforeCursor(), true)
}
