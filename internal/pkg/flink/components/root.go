package components

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func RootLayout(interactiveOutput tview.Primitive) *tview.Flex {
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(interactiveOutput, 0, 1, true)
	flex.SetBackgroundColor(tcell.ColorDefault)
	return flex
}
