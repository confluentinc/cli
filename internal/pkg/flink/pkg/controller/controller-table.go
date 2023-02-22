package controller

import (
	"os"
	"strings"

	clipboard "github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/olekukonko/tablewriter"
	"github.com/rivo/tview"
)

var table *tview.Table

type TableController struct {
	table        *tview.Table
	setData      func(data string)
	basic        func()
	separator    func()
	borders      func()
	selectRow    func()
	selectColumn func()
	selectCell   func()
	focus        func()
	onCtrlC      func()
}

type TableStyle struct {
	borders bool
}

func TableControllerInit(tableRef *tview.Table, appControler *ApplicationController) TableController {
	table = tableRef
	tableStyle := TableStyle{
		borders: false,
	}

	// Internal Functions
	printOutputModeTable := func(newData string) {
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
				table.SetCell(row, column, tableCell)
			}
		}
	}

	printInputModeTable := func(newData string) {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"OrderDate", "Region", "Rep", "Item", "Units", "UnitCost", "Total"})

		for _, tableRow := range strings.Split(newData, "\n") {
			row := strings.Split(tableRow, "|")
			table.Append(row)
		}

		table.Render() // Send output
	}

	// Actions
	setData := func(newData string) {
		table.Clear()

		if appControler.getView() == "outputMode" {
			// Print interactive text table
			printOutputModeTable(newData)
		} else {
			// Print raw text table
			printInputModeTable(newData)
		}

	}

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

	onCtrlC := func() {
		rowIndex, _ := table.GetSelection()
		columnCount := table.GetColumnCount()

		var row []string
		for i := 0; i < columnCount; i++ {
			row = append(row, table.GetCell(rowIndex, i).Text)
		}
		clipboardValue := strings.Join(row, ", ")

		clipboard.WriteAll(clipboardValue)
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
		table:        table,
		setData:      setData,
		basic:        basic,
		separator:    separator,
		borders:      borders,
		selectRow:    selectRow,
		selectColumn: selectColumn,
		selectCell:   selectCell,
		focus:        focus,
		onCtrlC:      onCtrlC,
	}
}
