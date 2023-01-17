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

	slidesParams := make(map[string]components.ExtraSlideParams)

	tableStyle := TableStyle{
		borders: false,
	}

	// Shortcuts text view showed on the botton
	shortcutsTV := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	// Functions used my the main component
	borders := func() {
		table := slidesParams["Home"].Table
		tableStyle.borders = !tableStyle.borders
		table.SetBorders(tableStyle.borders)
	}

	shortcutHighlighted := func(added, removed, remaining []string) {
		index, _ := strconv.Atoi(added[0])
		switch shortcuts[index].Text {
		case "Toggle Display Mode":
			borders()
		}
	}

	appInputCapture := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlT {
			borders()
			return nil
		}
		return event
	}

	// Instantiate InteractiveOutput Component
	title, params, InteractiveOutput := components.InteractiveOutput(func() {}, app)
	slidesParams[title] = params

	// Populate shortcuts from shortcuts array const
	for index, shortcut := range shortcuts {
		fmt.Fprintf(shortcutsTV, `[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text)
	}
	shortcutsTV.SetHighlightedFunc(shortcutHighlighted)
	shortcutsTV.Highlight("0")

	// Create the main layout.
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(InteractiveOutput, 0, 1, true).
		AddItem(shortcutsTV, 1, 1, false)

	// Shortcuts to navigate the slides.
	app.SetInputCapture(appInputCapture)

	// Start the application.
	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
