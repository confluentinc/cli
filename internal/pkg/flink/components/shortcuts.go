package components

import (
	"github.com/rivo/tview"
)

func Shortcuts() *tview.TextView {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)
}
