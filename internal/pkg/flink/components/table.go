package components

import (
	"github.com/rivo/tview"
)

func CreateTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 1)
	table.SetCell(0, 0, tview.NewTableCell("Starting SQL Client... "))
	table.SetBorder(true).SetTitle("Table")
	return table
}
