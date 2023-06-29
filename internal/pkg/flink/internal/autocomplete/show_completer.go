package autocomplete

import (
	prompt "github.com/confluentinc/go-prompt"
)

func ShowCompleter(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "SHOW CATALOGS;", Description: "Used to list catalogs"},
		{Text: "SHOW DATABASES;", Description: "Used to list databases"},
		{Text: "SHOW TABLES;", Description: "Lists all tables from current database"},
		{Text: "SHOW CURRENT CATALOG;", Description: "Displays the current catalog"},
		{Text: "SHOW CURRENT DATABASE;", Description: "Displays the current database"},
	}

	return SuggestFromPrefix(s, in.TextBeforeCursor())
}
