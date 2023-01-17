package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InputField demonstrates the InputField.
func InputField() *tview.InputField {
	return tview.NewInputField().
		SetText("SELECT * FROM ORDERS;").
		SetLabel("flinkSql[yellow]>>> ").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabelColor(tcell.ColorWhite)
}
