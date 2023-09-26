package components

import (
	"github.com/rivo/tview"
)

func InteractiveOutput(table *tview.Table, infoRow tview.Primitive, shortcuts *tview.TextView) tview.Primitive {
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexRow).
				AddItem(table, 0, 1, true),
			0, 1, false).
		AddItem(infoRow, 1, 1, false).
		AddItem(shortcuts, 1, 1, false)
}
