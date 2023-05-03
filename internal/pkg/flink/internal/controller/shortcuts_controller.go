package controller

import (
	"fmt"
	"strconv"

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
	KeyText string
	Text    string
}

// Keyboard shortcuts shown at the bottom.
var appShortcuts = []Shortcut{
	{KeyText: "Q", Text: "Quit"},
	//{KeyText: "N", Text: "Next Page"},
	{KeyText: "M", Text: "Toggle Result Mode"},
	{KeyText: "R", Text: "Toggle Auto Refresh"},
}

func (s *ShortcutsController) ShortcutHighlighted(added, removed, remaining []string) {
	index, _ := strconv.Atoi(added[0])
	action := s.tableController.GetActionForShortcut(appShortcuts[index].KeyText)
	if action != nil {
		action()
	}
}

func NewShortcutsController(shortcutsRef *tview.TextView, appController ApplicationControllerInterface, tableController TableControllerInterface) ShortcutsControllerInterface {
	for index, shortcut := range appShortcuts {
		fmt.Fprintf(shortcutsRef, `[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text)
	}

	return &ShortcutsController{shortcuts: shortcutsRef, appController: appController, tableController: tableController}
}
