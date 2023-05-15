package components

import (
	"fmt"

	"github.com/rivo/tview"
)

type Shortcut struct {
	KeyText string
	Text    string
}

// Keyboard shortcuts shown at the bottom.
var appShortcuts = []Shortcut{
	{KeyText: "Q", Text: "Quit"},
	{KeyText: "M", Text: "Toggle Result Mode"},
	{KeyText: "R", Text: "Toggle Auto Refresh"},
}

func Shortcuts() *tview.TextView {
	shortcutsRef := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	for index, shortcut := range appShortcuts {
		fmt.Fprintf(shortcutsRef, `[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text)
	}

	return shortcutsRef
}
