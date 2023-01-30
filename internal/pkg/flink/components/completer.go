package components

import (
	prompt "github.com/c-bata/go-prompt"
)

func completer(in prompt.Document) []prompt.Suggest {

	s := []prompt.Suggest{
		{Text: "SELECT", Description: "Select data from a database"},
		{Text: "INSERT", Description: "Add rows to a table"},
		{Text: "DESCRIBE", Description: "Describe the schema of a table or a view"},
		{Text: "SET", Description: "Set current database or catalog"},
	}

	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}
