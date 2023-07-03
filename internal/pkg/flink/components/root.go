package components

import (
	"github.com/rivo/tview"
)

func RootLayout(interactiveOutput tview.Primitive) *tview.Flex {
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(interactiveOutput, 0, 1, true)
	return flex
}
