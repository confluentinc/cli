package components

import (
	prompt "github.com/c-bata/go-prompt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ExtraSlideParams struct {
	Table *tview.Table
}

var input = tview.NewInputField().
	SetText("SELECT * FROM ORDERS;").
	SetLabel("flinkSql[yellow]>>> ").
	SetFieldBackgroundColor(tcell.ColorDefault).
	SetLabelColor(tcell.ColorWhite)

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

func InteractiveOutput(nextSlide func(), app *tview.Application) (title string, params ExtraSlideParams, content tview.Primitive) {

	InteractiveInput()

	list, table, _, selectRow, navigate := CreateTable(nextSlide, app)

	input.SetDoneFunc(func(key tcell.Key) {
		selectRow()
		navigate()
	})

	return "Home", ExtraSlideParams{table}, tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(input, 1, 1, true).
		AddItem(
			(tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(table, 0, 1, true)).
				AddItem(list, 10, 1, false),
			0, 1, false)
}

func getInput() *tview.InputField {
	return input
}
