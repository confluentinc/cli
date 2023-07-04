package components

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/samber/lo"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type TableView struct {
	RootLayout                           tview.Primitive
	table                                *tview.Table
	selectedRowIdx                       int
	tableLock                            sync.Mutex
	tableWidth                           int
	numRowsToScroll                      int
	isRowSelectionEnabled                bool
	materializedStatementResultsIterator types.MaterializedStatementResultsIterator
}

const (
	minColumnWidth int = 4 // min characters displayed in a column
)

func NewTableView() *TableView {
	t := &TableView{
		table:          createTable(),
		selectedRowIdx: -1,
	}
	t.table.SetSelectionChangedFunc(t.rowSelectionHandler)
	t.RootLayout = createTableView(t.table)
	return t
}

func createTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 1)
	table.SetBorder(true)
	return table
}

func (t *TableView) rowSelectionHandler(row, col int) {
	outOfBounds := row <= 0 || row >= t.table.GetRowCount()
	if outOfBounds || !t.isRowSelectionEnabled {
		return
	}

	if t.selectedRowIdx != -1 {
		stepsToMove := row - t.selectedRowIdx
		t.materializedStatementResultsIterator.Move(stepsToMove)
	}
	t.selectedRowIdx = row
}

func createTableView(table *tview.Table) *tview.Flex {
	shortcuts := Shortcuts()
	interactiveOutput := InteractiveOutput(table, shortcuts)
	rootLayout := RootLayout(interactiveOutput)
	return rootLayout
}

func (t *TableView) GetTable() *tview.Table {
	return t.table
}

func (t *TableView) GetSelectedRow() *types.StatementResultRow {
	return t.materializedStatementResultsIterator.Value()
}

func (t *TableView) RenderTable(tableTitle string, statementResults *types.MaterializedStatementResults, enableRowSelection bool) {
	t.tableLock.Lock()
	defer t.tableLock.Unlock()

	t.isRowSelectionEnabled = enableRowSelection
	// reset the iterator and the selected idx
	t.selectedRowIdx = -1
	t.materializedStatementResultsIterator = statementResults.Iterator(true)

	t.table.SetTitle(tableTitle)
	t.renderData(statementResults)
	t.selectLastRow()
}

func (t *TableView) renderData(statementResults *types.MaterializedStatementResults) {
	t.table.Clear()

	_, _, tableWidth, _ := t.table.GetInnerRect()
	t.tableWidth = tableWidth
	columnWidths := statementResults.GetMaxWidthPerColumn()
	truncatedColumnWidths := results.GetTruncatedColumnWidths(columnWidths, tableWidth)

	// Print header
	for colIdx, column := range statementResults.GetHeaders() {
		tableCell := tview.NewTableCell(column).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false).
			SetMaxWidth(truncatedColumnWidths[colIdx])
		t.table.SetCell(0, colIdx, tableCell)
	}

	// Print content
	statementResults.ForEach(t.fillTable(truncatedColumnWidths))

	// add callback function for after draw (gets triggered on any render event, such as screen size update)
	t.table.SetDrawFunc(t.resizeTable(columnWidths))
}

func (t *TableView) fillTable(truncatedColumnWidths []int) func(rowIdx int, row *types.StatementResultRow) {
	return func(rowIdx int, row *types.StatementResultRow) {
		for colIdx, field := range row.Fields {
			tableCell := tview.NewTableCell(tview.Escape(field.ToString())).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(truncatedColumnWidths[colIdx])
			t.table.SetCell(rowIdx+1, colIdx, tableCell)
		}
	}
}

func (t *TableView) resizeTable(columnWidths []int) func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	return func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		// check if the table width has changed
		newX, newY, newWidth, newHeight := t.table.GetInnerRect()
		hasTableWidthChanged := t.tableWidth != newWidth
		t.tableWidth = newWidth
		// minus 2 because of the header row and because we want to go to the first row we can still see
		t.numRowsToScroll = newHeight - 2
		if !hasTableWidthChanged {
			return newX, newY, newWidth, newHeight
		}

		// check if space needed fits screen, if it doesn't truncate the column
		truncatedColumnWidths := results.GetTruncatedColumnWidths(columnWidths, newWidth)
		for rowIdx := 0; rowIdx < t.table.GetRowCount(); rowIdx++ {
			for colIdx := 0; colIdx < t.table.GetColumnCount(); colIdx++ {
				t.table.GetCell(rowIdx, colIdx).SetMaxWidth(lo.Max([]int{truncatedColumnWidths[colIdx], minColumnWidth}))
			}
		}
		return newX, newY, newWidth, newHeight
	}
}

func (t *TableView) selectLastRow() {
	t.table.SetSelectable(t.isRowSelectionEnabled, false).Select(t.table.GetRowCount()-1, 0)
	t.table.ScrollToEnd()
}

func (t *TableView) FastScrollUp() {
	rowToSelect := lo.Max([]int{1, t.selectedRowIdx - t.numRowsToScroll})
	t.table.Select(rowToSelect, 0)
}

func (t *TableView) FastScrollDown() {
	rowToSelect := lo.Min([]int{t.table.GetRowCount() - 1, t.selectedRowIdx + t.numRowsToScroll})
	t.table.Select(rowToSelect, 0)
}
