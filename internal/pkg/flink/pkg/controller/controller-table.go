package controller

import (
	"strings"

	clipboard "github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TableController struct {
	table           *tview.Table
	tableStyle      TableStyle
	appController   *ApplicationController
	InputController *InputController
	store           Store
}

type TableStyle struct {
	borders bool
}

func (t *TableController) setData(newData string) {
	t.table.Clear()

	// Print interactive text table
	for row, line := range strings.Split(newData, "\n") {
		for column, cell := range strings.Split(line, "|") {
			color := tcell.ColorWhite
			if row == 0 {
				color = tcell.ColorYellow
			} else if column == 0 {
				color = tcell.ColorDarkCyan
			}
			align := tview.AlignLeft
			if row == 0 {
				align = tview.AlignCenter
			} else if column == 0 || column >= 4 {
				align = tview.AlignRight
			}
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(align).
				SetSelectable(row != 0 && column != 0)
			if column >= 1 && column <= 3 {
				tableCell.SetExpansion(1)
			}
			t.table.SetCell(row, column, tableCell)
		}
	}
}

func (t *TableController) handleCellEvent(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEscape {
		t.appController.toggleOutputMode()
		t.appController.suspendOutputMode(t.InputController.RunInteractiveInput)
		t.fetchDataAndPrintTable()
		return nil
	}

	return event
}

func (t *TableController) borders() {
	t.tableStyle.borders = !t.tableStyle.borders
	t.table.SetBorders(t.tableStyle.borders)
}

func (t *TableController) selectRow() {
	t.table.SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ')
}

// TODO: these look unused
//func (t *TableController) basic() {
//	t.table.SetBorders(false).
//		SetSelectable(false, false).
//		SetSeparator(' ')
//}

//func (t *TableController) separator() {
//	t.table.SetBorders(false).
//		SetSelectable(false, false).
//		SetSeparator(tview.Borders.Vertical)
//}

//func (t *TableController) selectColumn() {
//	t.table.SetBorders(false).
//		SetSelectable(false, true).
//		SetSeparator(' ')
//}
//
//func (t *TableController) selectCell() {
//	t.table.SetBorders(false).
//		SetSelectable(true, true).
//		SetSeparator(' ')
//}

func (t *TableController) focus() {
	t.selectRow()
	t.appController.app.SetFocus(t.table)
}

func (t *TableController) onCtrlC() {
	rowIndex, _ := t.table.GetSelection()
	columnCount := t.table.GetColumnCount()

	var row []string
	for i := 0; i < columnCount; i++ {
		row = append(row, t.table.GetCell(rowIndex, i).Text)
	}
	clipboardValue := strings.Join(row, ", ")

	clipboard.WriteAll(clipboardValue)
}

// Function to handle shortcuts and keybindings for the whole app
func (a *TableController) appInputCapture(event *tcell.EventKey) *tcell.EventKey {

	if event.Key() == tcell.KeyCtrlT {
		a.borders()
		return nil
		// TODO we have to actually go forward and backwards and not only go to the next mock
	} else if event.Key() == tcell.KeyCtrlN || event.Key() == tcell.KeyCtrlP {
		data := a.store.FetchData("")
		a.setData(data)
		return nil
	} else if event.Key() == tcell.KeyCtrlC {
		a.onCtrlC()
		return nil
	}
	return event

}

func (a *TableController) PrintTable(data string) {
	a.setData(data)
}

func (a *TableController) fetchDataAndPrintTable() {
	data := a.store.FetchData("")
	a.PrintTable(data)
}

func NewTableController(tableRef *tview.Table, store Store, appController *ApplicationController) *TableController {
	controller := &TableController{
		table:         tableRef,
		store:         store,
		appController: appController,
	}
	return controller
}
