package autocomplete

import (
	prompt "github.com/confluentinc/go-prompt"
)

func ExamplesCompleter(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "CREATE TABLE ", Description: "Register a table/view/function into current or specified Catalog"},
		{Text: "ALTER TABLE ", Description: "Modify a registered table/view/function definition in the Catalog"},
		{Text: "DESCRIBE ", Description: "Describe the schema of a table or a view"},
		{Text: "INSERT INTO ", Description: "Add rows to a table"},
		{Text: "USE ", Description: "Used to set the current database or catalog"},
		{Text: "RESET;", Description: "Used to reset the configuration to the default"},
		{Text: "SELECT ", Description: "Select data from a database"},
		{Text: "SET ", Description: "Used to modify the configuration or list the configuration"},
	}

	return SuggestFromPrefix(s, in.TextBeforeCursor())
}
