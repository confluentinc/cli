package autocomplete

import (
	prompt "github.com/confluentinc/go-prompt"
)

func ExamplesCompleter(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "CREATE TABLE ", Description: "Register a table/view/function into current or specified Catalog"},
		{Text: "DROP TABLE ", Description: "Drop a table/view/function from current or specified Catalog or the catalog itself"},
		{Text: "ALTER TABLE ", Description: "Modify a registered table/view/function definition in the Catalog"},
		{Text: "INSERT INTO ", Description: "Add rows to a table"},
		{Text: "DESCRIBE ", Description: "Describe the schema of a table or a view"},
		{Text: "EXPLAIN CHANGELOG_MODE, PLAN_ADVICE SELECT...;", Description: "Used to explain the logical and optimized query plans of a query or an INSERT statement"},
		{Text: "USE ", Description: "Used to set the current database or catalog"},
		{Text: "SHOW TABLES;", Description: "Used to list different flink objects such as catalogs, databases, tables and more."},
		{Text: "RESET;", Description: "Used to reset the configuration to the default"},
		{Text: "SELECT ", Description: "Select data from a database"},
		{Text: "SET ", Description: "Used to modify the configuration or list the configuration"},
	}

	return SuggestFromPrefix(s, in.TextBeforeCursor())
}
