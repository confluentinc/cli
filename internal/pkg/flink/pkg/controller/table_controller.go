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
	appController   ApplicationControllerInterface
	InputController *InputController
	store           StoreInterface
}

type TableStyle struct {
	borders bool
}

func (t *TableController) setData(newData *StatementResult) {
	t.table.Clear()

	// Print header
	for col, header := range newData.Columns {
		tableCell := tview.NewTableCell(header).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter).
			SetSelectable(false)
		t.table.SetCell(0, col, tableCell)
	}
	// Print content
	for row, line := range newData.Rows {
		for column, cell := range line {
			color := tcell.ColorWhite
			if column == 0 {
				color = tcell.ColorDarkCyan
			}
			align := tview.AlignLeft
			if column == 0 || column >= 4 {
				align = tview.AlignRight
			}
			tableCell := tview.NewTableCell(cell).
				SetTextColor(color).
				SetAlign(align).
				SetSelectable(column != 0)
			if column >= 1 && column <= 3 {
				tableCell.SetExpansion(1)
			}
			t.table.SetCell(row+1, column, tableCell)
		}
	}
}

func (t *TableController) handleCellEvent(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEscape {
		// Here we suspend outpude mode/tview aand run the interactive input again
		t.appController.SuspendOutputMode(t.InputController.RunInteractiveInput)

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
	t.appController.TView().SetFocus(t.table)
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
		// We send select so we can get the next mock
		data, err := a.store.ProcessStatement("select ;")
		if err == nil {
			a.setData(data)
		}
		return nil
	} else if event.Key() == tcell.KeyCtrlC {
		a.onCtrlC()
		return nil
	}
	return event

}

// This function will be changed when we actually use tview
func (a *TableController) setDataAndFocus(statementResult *StatementResult) {
	a.setData(statementResult)
	a.focus()
}

func NewTableController(tableRef *tview.Table, store StoreInterface, appController ApplicationControllerInterface) *TableController {
	controller := &TableController{
		table:         tableRef,
		store:         store,
		appController: appController,
	}
	return controller
}
