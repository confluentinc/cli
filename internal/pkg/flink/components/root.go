package components

import (
	"github.com/rivo/tview"
)

func RootLayout(interactiveOutput tview.Primitive) *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(interactiveOutput, 0, 1, true)
}
