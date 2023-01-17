package main

import (
	"fmt"
	"strconv"

	components "github.com/confluentinc/flink-sql-client/components"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Shortcut struct {
	Key     tcell.Key
	KeyText string
	Text    string
}

type TableStyle struct {
	borders bool
}

// Keyboard shortcuts shown at the bottom.
var shortcuts = []Shortcut{
	{Key: tcell.KeyCtrlS, KeyText: "S", Text: "Smart Completion"},
	{Key: tcell.KeyCtrlM, KeyText: "M", Text: "Multiline"},
	{Key: tcell.KeyCtrlM, KeyText: "T", Text: "Toggle Display Mode"},
	{Key: tcell.KeyCtrlN, KeyText: "N", Text: "Next slide"},
	{Key: tcell.KeyCtrlP, KeyText: "P", Text: "Previous slide"}}

// Tview application.
var app = tview.NewApplication()

func main() {
	tableController := TableControllerInit(components.CreateTable())
	InputControllerInit(components.InputField())
	ApplicationControllerInit(tableController)

	// Shortcuts text view showed on the botton
	shortcutsTV := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	// Functions used my the main component
	shortcutHighlighted := func(added, removed, remaining []string) {
		index, _ := strconv.Atoi(added[0])
		switch shortcuts[index].Text {
		case "Toggle Display Mode":
			tableController.borders()
		}
	}

	// Instantiate components
	components.InteractiveInput()
	interactiveOutput := components.InteractiveOutput(input, table)

	// Populate shortcuts from shortcuts array const
	for index, shortcut := range shortcuts {
		fmt.Fprintf(shortcutsTV, `[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text)
	}
	shortcutsTV.SetHighlightedFunc(shortcutHighlighted)
	shortcutsTV.Highlight("0")

	// Create the main layout.
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(interactiveOutput, 0, 1, true).
		AddItem(shortcutsTV, 1, 1, false)

	// Start the application.
	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
