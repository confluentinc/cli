package autocomplete

import (
	prompt "github.com/c-bata/go-prompt"
)

func examplesCompleter(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "CREATE TABLE Orders (product STRING, amount INT);", Description: "Register a table/view/function into current or specified Catalog"},
		{Text: "DROP TABLE Orders;", Description: "Drop a table/view/function from current or specified Catalog or the catalog itself"},
		{Text: "ALTER TABLE Orders RENAME TO NewOrders;", Description: "Modify a registered table/view/function definition in the Catalog"},
		{Text: "INSERT INTO Orders VALUES ('pen', 2);", Description: "Add rows to a table"},
		{Text: "DESCRIBE Orders;", Description: "Describe the schema of a table or a view"},
		{Text: "EXPLAIN CHANGELOG_MODE, PLAN_ADVICE SELECT...;", Description: "Used to explain the logical and optimized query plans of a query or an INSERT statement"},
		{Text: "USE db1;", Description: "Used to set the current database or catalog"},
		{Text: "SHOW TABLES;", Description: "Used to list different flink objects such as catalogs, databases, tables and more."},
		{Text: "SET 'table.local-time-zone;' = 'Europe/Berlin';", Description: "Used to modify the configuration or list the configuration"},
		{Text: "RESET;", Description: "Used to reset the configuration to the default"},
		{Text: "SELECT * FROM Orders WHERE amount = 2;", Description: "Select data from a database"},
	}

	return SuggestFromPrefix(s, in.TextBeforeCursor())
}
