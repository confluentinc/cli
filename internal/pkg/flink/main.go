package main

import (
	"fmt"
	"strconv"

	components "github.com/confluentinc/flink-sql-client/components"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Slide func(nextSlide func(), appRef *tview.Application) (title string, params components.ExtraSlideParams, content tview.Primitive)

type Shortcut struct {
	Key     tcell.Key
	KeyText string
	Text    string
}

type TableStyle struct {
	borders bool
}

// Keyboard shortcuts shown on the bottom.
var shortcuts = []Shortcut{
	{Key: tcell.KeyCtrlS, KeyText: "S", Text: "Smart Completion"},
	{Key: tcell.KeyCtrlM, KeyText: "M", Text: "Multiline"},
	{Key: tcell.KeyCtrlM, KeyText: "T", Text: "Toggle Display Mode"},
	{Key: tcell.KeyCtrlN, KeyText: "N", Text: "Next slide"},
	{Key: tcell.KeyCtrlP, KeyText: "P", Text: "Previous slide"}}

// The application.
var app = tview.NewApplication()

// Starting point for the presentation.
func main() {

	// The presentation slides.
	slides := []Slide{
		components.TableWithInput,
		components.Cover,
		components.InputField,
		components.Table,
		components.Introduction,
		components.HelloWorld,
		components.Form,
		components.TextView1,
		components.TextView2,
		components.TreeView,
		components.Flex,
		components.Grid,
		components.Colors,
		components.End,
	}
	slidesParams := make(map[string]components.ExtraSlideParams)
	pages := tview.NewPages()

	tableStyle := TableStyle{
		borders: false,
	}

	// The bottom row has some info on where we are.
	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false).
		SetHighlightedFunc(func(added, removed, remaining []string) {
			pages.SwitchToPage(added[0])
		})

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
	previousSlide := func() {
		slide, _ := strconv.Atoi(info.GetHighlights()[0])
		slide = (slide - 1 + len(slides)) % len(slides)
		info.Highlight(strconv.Itoa(slide)).
			ScrollToHighlight()
	}
	nextSlide := func() {
		slide, _ := strconv.Atoi(info.GetHighlights()[0])
		slide = (slide + 1) % len(slides)
		info.Highlight(strconv.Itoa(slide)).
			ScrollToHighlight()
	}
	shortcutHighlighted := func(added, removed, remaining []string) {
		index, _ := strconv.Atoi(added[0])
		switch shortcuts[index].Text {
		case "Toggle Display Mode":
			borders()
		case "Next slide":
			nextSlide()
		case "Previous slide":
			previousSlide()
		}
	}
	appInputCapture := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlT {
			borders()
			return nil
		} else if event.Key() == tcell.KeyCtrlN {
			nextSlide()
			return nil
		} else if event.Key() == tcell.KeyCtrlP {
			previousSlide()
			return nil
		}
		return event
	}

	// Populating info text view and pages with slides
	for index, slide := range slides {
		title, params, primitive := slide(nextSlide, app)
		slidesParams[title] = params
		pages.AddPage(strconv.Itoa(index), primitive, true, index == 0)
		fmt.Fprintf(info, `%d ["%d"][darkcyan]%s[white][""]  `, index+1, index, title)
	}

	// Populating shortcuts from shortcuts array const
	for index, shortcut := range shortcuts {
		fmt.Fprintf(shortcutsTV, `[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text)
	}

	info.Highlight("0")
	shortcutsTV.SetHighlightedFunc(shortcutHighlighted)
	shortcutsTV.Highlight("0")

	// Create the main layout.
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		//AddItem(info, 1, 1, false).
		AddItem(shortcutsTV, 1, 1, false)

	// Shortcuts to navigate the slides.
	app.SetInputCapture(appInputCapture)

	// Start the application.
	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}
