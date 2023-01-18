package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func InputField() *tview.InputField {
	return tview.NewInputField().
		SetLabel("flinkSql[yellow]>>> ").
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetLabelColor(tcell.ColorWhite)
}
