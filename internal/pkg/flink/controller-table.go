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
	focus        func()
}

type TableStyle struct {
	borders bool
}

func TableControllerInit(tableRef *tview.Table, appControler *ApplicationController) TableController {
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

	selectRow := func() {
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

	focus := func() {
		selectRow()
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
		selectRow:    selectRow,
		selectColumn: selectColumn,
		selectCell:   selectCell,
		focus:        focus,
	}
}
