package components

import (
	prompt "github.com/c-bata/go-prompt"

	"github.com/rivo/tview"
)

type ExtraSlideParams struct {
	Table *tview.Table
	Input *tview.InputField
}

func completer(in prompt.Document) []prompt.Suggest {
	prompt.NewStdoutWriter().WriteRawStr("completer")

	s := []prompt.Suggest{
		{Text: "SELECT", Description: "Select data from a database"},
		{Text: "INSERT", Description: "Add rows to a table"},
		{Text: "DESCRIBE", Description: "Describe the schema of a table or a view"},
		{Text: "SET", Description: "Set current database or catalog"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func InteractiveOutput(input *tview.InputField, table *tview.Table) tview.Primitive {

	InteractiveInput()

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(input, 1, 1, true).
		AddItem(
			(tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(table, 0, 1, true)),
			0, 1, false)
}
