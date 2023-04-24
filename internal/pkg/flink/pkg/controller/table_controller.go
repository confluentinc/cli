package controller

import (
	"strings"
	"unicode"

	"github.com/confluentinc/flink-sql-client/pkg/types"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TableControllerInterface interface {
	AppInputCapture(event *tcell.EventKey) *tcell.EventKey
	SetDataAndFocus(statement types.ProcessedStatement)
	SetInputController(inputController InputControllerInterface)
}

type TableController struct {
	table           *tview.Table
	appController   ApplicationControllerInterface
	InputController InputControllerInterface
	store           StoreInterface
	results         types.ProcessedStatement
}

func NewTableController(tableRef *tview.Table, store StoreInterface, appController ApplicationControllerInterface) TableControllerInterface {
	controller := &TableController{
		table:         tableRef,
		store:         store,
		appController: appController,
	}
	return controller
}

func (t *TableController) SetInputController(inputController InputControllerInterface) {
	t.InputController = inputController
}

// Function to handle shortcuts and keybindings for TView
func (t *TableController) AppInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		char := unicode.ToUpper(event.Rune())
		switch char {
		case 'Q':
			t.appController.SuspendOutputMode(t.InputController.RunInteractiveInput)
			return nil
		case 'N':
			// fetch new page
			newResults, err := t.store.FetchStatementResults(t.results)
			if err != nil {
				// todo: handle error when we know which kinds of error to expect from backend - we'll have a ticket
			}
			t.SetDataAndFocus(*newResults)
			return nil
		}
	} else {
		switch event.Key() {
		case tcell.KeyCtrlC:
			t.onCtrlC()
			return nil
		case tcell.KeyEscape:
			t.appController.SuspendOutputMode(t.InputController.RunInteractiveInput)
			return nil
		case tcell.KeyCtrlQ:
			t.appController.SuspendOutputMode(t.InputController.RunInteractiveInput)
			return nil
		}
	}
	return event
}

// This function will be changed when we actually use tview
func (t *TableController) SetDataAndFocus(statement types.ProcessedStatement) {
	t.results = statement
	t.setData(statement.StatementResults)
	t.focus()
}

func (t *TableController) setData(statementResults *types.StatementResults) {
	if statementResults == nil {
		return
	}
	t.table.Clear()
	// Print header
	for colIdx, column := range statementResults.Headers {
		tableCell := tview.NewTableCell(column).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		t.table.SetCell(0, colIdx, tableCell)
	}

	// Print content
	for rowIdx, row := range statementResults.Rows {
		for colIdx, field := range row.Fields {
			color := tcell.ColorWhite
			tableCell := tview.NewTableCell(tview.Escape(field.Format(nil))).
				SetTextColor(color).
				SetAlign(tview.AlignLeft).
				SetSelectable(true)
			t.table.SetCell(rowIdx+1, colIdx, tableCell)
		}
	}
}

func (t *TableController) selectRow() {
	t.table.SetBorders(false).
		SetSelectable(true, false).
		SetSeparator(' ')
}

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
