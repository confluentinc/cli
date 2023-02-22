package controller

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ShortcutsController struct {
}

type Shortcut struct {
	Key     tcell.Key
	KeyText string
	Text    string
}

// Keyboard shortcuts shown at the bottom.
var appShortcuts = []Shortcut{
	{Key: tcell.KeyCtrlQ, KeyText: "Q", Text: "Quit"},
	{Key: tcell.KeyCtrlS, KeyText: "S", Text: "Smart Completion"},
	{Key: tcell.KeyCtrlM, KeyText: "M", Text: "Interactive Output ON/OFF"},
	{Key: tcell.KeyCtrlT, KeyText: "T", Text: "Toggle Display Mode"},
	{Key: tcell.KeyCtrlT, KeyText: "N", Text: "Next Page"},
	{Key: tcell.KeyCtrlT, KeyText: "P", Text: "Prev Page"},
}

var shortcuts *tview.TextView

func ShortcutsControllerInit(shortcutsRef *tview.TextView, tableController TableController, appControler *ApplicationController) ShortcutsController {
	shortcuts = shortcutsRef

	for index, shortcut := range appShortcuts {
		fmt.Fprintf(shortcuts, `[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text)
	}

	shortcutHighlighted := func(added, removed, remaining []string) {
		index, _ := strconv.Atoi(added[0])
		switch appShortcuts[index].Text {
		case "Toggle Display Mode":
			tableController.borders()
		case "Quit":
			appControler.exitApplication()
		}
	}

	shortcuts.SetHighlightedFunc(shortcutHighlighted)

	return ShortcutsController{}
}
