package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var table *tview.Table

type TableController struct {
	basic        func()
	separator    func()
	borders      func()
	selectRow    func()
	selectColumn func()
	selectCell   func()
	navigate     func()
}

type TableStyle struct {
	borders bool
}

var selectRow = func() {
	table.SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ')
}

var navigate = func() {
	app.SetFocus(table)
}

func TableControllerInit(tableRef *tview.Table) TableController {
	table = tableRef
	tableStyle := TableStyle{
		borders: false,
	}

	// Actions
	basic := func() {
		table.SetBorders(false).
			SetSelectable(false, false).
			SetSeparator(' ')
	}

	separator := func() {
		table.SetBorders(false).
			SetSelectable(false, false).
			SetSeparator(tview.Borders.Vertical)
	}

	borders := func() {
		tableStyle.borders = !tableStyle.borders
		table.SetBorders(tableStyle.borders)
	}

	selectRow2 := func() {
		table.SetBorders(false).
			SetSelectable(true, false).
			SetSeparator(' ')
	}

	selectColumn := func() {
		table.SetBorders(false).
			SetSelectable(false, true).
			SetSeparator(' ')
	}

	selectCell := func() {
		table.SetBorders(false).
			SetSelectable(true, true).
			SetSeparator(' ')
	}

	navigate2 := func() {
		app.SetFocus(table)
	}

	// Configure table
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			app.SetFocus(input)
			return nil
		}

		return event
	})

	return TableController{
		basic:        basic,
		separator:    separator,
		borders:      borders,
		selectRow:    selectRow2,
		selectColumn: selectColumn,
		selectCell:   selectCell,
		navigate:     navigate2,
	}
}
