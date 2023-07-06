package components

import (
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/samber/lo"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type TableViewInterface interface {
	GetFocusableElement() *tview.Table
	GetRoot() tview.Primitive
	GetSelectedRow() *types.StatementResultRow
	RenderTable(tableTitle string, statementResults *types.MaterializedStatementResults, isAutoRefreshRunning bool)
	FastScrollUp()
	FastScrollDown()
}

type TableView struct {
	rootLayout      tview.Primitive
	table           *tview.Table
	tableLock       sync.Mutex
	tableWidth      int
	numRowsToScroll int
}

const (
	minColumnWidth            int = 4 // min characters displayed in a column
	ExitTableViewShortcut         = "Q"
	ToggleAutoRefreshShortcut     = "P"
	ToggleTableModeShortcut       = "M"
	JumpUpShortcut                = "U"
	JumpDownShortcut              = "D"
)

func NewTableView() TableViewInterface {
	return &TableView{
		table: createTable(),
	}
}

func createTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 1)
	table.SetBorder(true)
	return table
}

func createTableView(table *tview.Table, shortcuts *tview.TextView) *tview.Flex {
	interactiveOutput := InteractiveOutput(table, shortcuts)
	rootLayout := RootLayout(interactiveOutput)
	return rootLayout
}

func (t *TableView) GetFocusableElement() *tview.Table {
	return t.table
}

func (t *TableView) GetRoot() tview.Primitive {
	return t.rootLayout
}

func (t *TableView) GetSelectedRow() *types.StatementResultRow {
	cell := t.table.GetCell(t.getSelectedRowIdx(), 0)
	if cell == nil {
		return nil
	}

	row, ok := cell.GetReference().(*types.StatementResultRow)
	if !ok {
		return nil
	}
	return row
}

func (t *TableView) getSelectedRowIdx() int {
	rowIdx, _ := t.table.GetSelection()
	return rowIdx
}

func (t *TableView) RenderTable(tableTitle string, statementResults *types.MaterializedStatementResults, isAutoRefreshRunning bool) {
	t.tableLock.Lock()
	defer t.tableLock.Unlock()

	t.rootLayout = createTableView(t.table, NewShortcuts(t.getTableShortcuts(statementResults, isAutoRefreshRunning)))
	t.table.SetTitle(tableTitle)
	t.renderData(statementResults)
	t.selectLastRow(!isAutoRefreshRunning)
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
				SetMaxWidth(truncatedColumnWidths[colIdx]).
				SetReference(row)
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

func (t *TableView) selectLastRow(enableRowSelection bool) {
	t.table.SetSelectable(enableRowSelection, false).Select(t.table.GetRowCount()-1, 0)
	t.table.ScrollToEnd()
}

func (t *TableView) FastScrollUp() {
	selectedRow, _ := t.table.GetSelection()
	rowToSelect := lo.Max([]int{1, selectedRow - t.numRowsToScroll})
	t.table.Select(rowToSelect, 0)
}

func (t *TableView) FastScrollDown() {
	selectedRow, _ := t.table.GetSelection()
	rowToSelect := lo.Min([]int{t.table.GetRowCount() - 1, selectedRow + t.numRowsToScroll})
	t.table.Select(rowToSelect, 0)
}

func (t *TableView) getTableShortcuts(statementResults *types.MaterializedStatementResults, isAutoRefreshRunning bool) []types.Shortcut {
	mode := "Show table"
	if statementResults.IsTableMode() {
		mode = "Show changelog"
	}
	playPause := "Play"
	if isAutoRefreshRunning {
		playPause = "Pause"
	}
	return []types.Shortcut{
		{KeyText: ExitTableViewShortcut, Text: "Quit"},
		{KeyText: ToggleTableModeShortcut, Text: mode},
		{KeyText: ToggleAutoRefreshShortcut, Text: playPause},
		{KeyText: fmt.Sprintf("%s/%s", JumpUpShortcut, JumpDownShortcut), Text: "Jump up/down"},
	}
}
