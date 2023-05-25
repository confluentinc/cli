package autocomplete

import (
	prompt "github.com/confluentinc/go-prompt"
)

func ShowCompleter(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "SHOW CATALOGS;", Description: "Used to list catalogs"},
		{Text: "SHOW DATABASES;", Description: "Used to list databases"},
		{Text: "SHOW TABLES;", Description: "Lists all tables from current database"},
		{Text: "SHOW TABLES FROM ;", Description: "Lists all tables from a database"},
		{Text: "SHOW COLUMNS FROM ;", Description: "Lists all columns of a table"},
		{Text: "SHOW VIEWS;", Description: "Lists all views"},
		{Text: "SHOW CURRENT CATALOG;", Description: "Displays the current catalog"},
		{Text: "SHOW CURRENT DATABASE;", Description: "Displays the current database"},
		{Text: "SHOW CREATE TABLE;", Description: "Displays the CREATE TABLE statement for a table"},
		{Text: "SHOW CREATE VIEW;", Description: "Displays the CREATE VIEW statement for a view"},
	}

	return SuggestFromPrefix(s, in.TextBeforeCursor())
}
