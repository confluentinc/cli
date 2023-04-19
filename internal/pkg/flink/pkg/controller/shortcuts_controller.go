package controller

import (
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ShortcutsControllerInterface interface {
	ShortcutHighlighted(added, removed, remaining []string)
}

type ShortcutsController struct {
	appController   ApplicationControllerInterface
	tableController TableControllerInterface
	shortcuts       *tview.TextView
}

type Shortcut struct {
	Key     tcell.Key
	KeyText string
	Text    string
}

// Keyboard shortcuts shown at the bottom.
var appShortcuts = []Shortcut{
	{Key: tcell.KeyCtrlQ, KeyText: "Q", Text: "Quit"},
	{Key: tcell.KeyCtrlT, KeyText: "N", Text: "Next Page"},
}

func (s *ShortcutsController) ShortcutHighlighted(added, removed, remaining []string) {
	index, _ := strconv.Atoi(added[0])
	switch appShortcuts[index].Text {
	case "Quit":
		// Todo - go back to input mode
		s.appController.ExitApplication()
	}
}

func NewShortcutsController(shortcutsRef *tview.TextView, appController ApplicationControllerInterface, tableController TableControllerInterface) ShortcutsControllerInterface {
	for index, shortcut := range appShortcuts {
		fmt.Fprintf(shortcutsRef, `[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text)
	}

	return &ShortcutsController{shortcuts: shortcutsRef, appController: appController, tableController: tableController}
}
