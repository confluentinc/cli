package components

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/samber/lo"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type TableViewInterface interface {
	Init()
	GetFocusableElement() *tview.Table
	GetRoot() tview.Primitive
	GetSelectedRow() *types.StatementResultRow
	RenderTable(tableTitle string, statementResults *types.MaterializedStatementResults, lastRefreshTimestamp *time.Time, fetchState types.FetchState)
	JumpUp()
	JumpDown()
}

type TableView struct {
	rootLayout   tview.Primitive
	table        *tview.Table
	tableLock    sync.Mutex
	tableWidth   int
	columnWidths []int
	infoBar      *TableInfoBar
}

const (
	numPaddingRows                = 1
	minColumnWidth            int = 4 // min characters displayed in a column
	ExitTableViewShortcut         = "Q"
	ToggleAutoRefreshShortcut     = "P"
	ToggleTableModeShortcut       = "M"
	JumpUpShortcut                = "U"
	JumpDownShortcut              = "D"
)

func NewTableView() TableViewInterface {
	return &TableView{}
}

func (t *TableView) Init() {
	t.infoBar = NewTableInfoBar()

	t.table = tview.NewTable().SetFixed(1, 1)
	t.table.SetBorder(true)
	t.table.SetSelectionChangedFunc(func(row, column int) {
		if t.isValidRowIdx(row) {
			t.updateInfoBar()
		}
	})
	t.table.SetDrawFunc(t.tableAfterDrawHandler())
}

func (t *TableView) isValidRowIdx(row int) bool {
	// when table is empty do nothing
	if t.getLastRowIdx() <= 0 {
		return false
	}
	// when table is not empty but we stepped out of bounds
	if row <= 0 {
		t.table.ScrollToBeginning()
		t.table.Select(1, 0)
		return false
	}
	if row >= t.table.GetRowCount()-numPaddingRows {
		t.table.ScrollToEnd()
		t.table.Select(t.getLastRowIdx(), 0)
		return false
	}

	return true
}

func (t *TableView) getLastRowIdx() int {
	return t.table.GetRowCount() - 1 - numPaddingRows
}

func (t *TableView) updateInfoBar() {
	if rowsSelectable, _ := t.table.GetSelectable(); rowsSelectable {
		t.infoBar.SetRowInfo(t.getSelectedRowIdx(), t.getLastRowIdx())
		return
	}
	t.infoBar.SetRowInfo(0, 0)
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

func (t *TableView) RenderTable(tableTitle string, statementResults *types.MaterializedStatementResults, lastRefreshTimestamp *time.Time, fetchState types.FetchState) {
	t.tableLock.Lock()
	defer t.tableLock.Unlock()

	t.infoBar.SetLastRefreshTimestamp(lastRefreshTimestamp)
	t.infoBar.SetFetchState(fetchState)
	t.createTableView(NewShortcuts(t.getTableShortcuts(statementResults, fetchState)))
	t.setTableAndColumnWidths(statementResults)

	t.table.SetTitle(tableTitle)
	t.renderData(statementResults)
	t.selectLastRow(fetchState != types.Running)
}

func (t *TableView) createTableView(shortcuts *tview.TextView) {
	interactiveOutput := InteractiveOutput(t.table, t.infoBar.GetView(), shortcuts)
	t.rootLayout = RootLayout(interactiveOutput)
}

func (t *TableView) setTableAndColumnWidths(statementResults *types.MaterializedStatementResults) {
	_, _, tableWidth, _ := t.table.GetInnerRect()
	t.tableWidth = tableWidth
	t.columnWidths = statementResults.GetMaxWidthPerColumn()
}

func (t *TableView) tableAfterDrawHandler() func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
	return func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		// check if the table width has changed
		newX, newY, newWidth, newHeight := t.table.GetInnerRect()
		hasTableWidthChanged := t.tableWidth != newWidth
		t.tableWidth = newWidth
		if hasTableWidthChanged {
			t.truncateTableColumns()
		}

		return newX, newY, newWidth, newHeight
	}
}

func (t *TableView) truncateTableColumns() {
	truncatedColumnWidths := results.GetTruncatedColumnWidths(t.columnWidths, t.tableWidth)
	for rowIdx := 0; rowIdx < t.table.GetRowCount(); rowIdx++ {
		for colIdx := 0; colIdx < t.table.GetColumnCount(); colIdx++ {
			t.table.GetCell(rowIdx, colIdx).SetMaxWidth(lo.Max([]int{truncatedColumnWidths[colIdx], minColumnWidth}))
		}
	}
}

func (t *TableView) renderData(statementResults *types.MaterializedStatementResults) {
	t.table.Clear()

	truncatedColumnWidths := results.GetTruncatedColumnWidths(t.columnWidths, t.tableWidth)

	t.addHeaderRow(statementResults, truncatedColumnWidths)
	t.addContentRows(statementResults, truncatedColumnWidths)
	t.addPaddingRows()
}

func (t *TableView) addHeaderRow(statementResults *types.MaterializedStatementResults, truncatedColumnWidths []int) {
	for colIdx, column := range statementResults.GetHeaders() {
		tableCell := tview.NewTableCell(column).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false).
			SetMaxWidth(truncatedColumnWidths[colIdx])
		t.table.SetCell(0, colIdx, tableCell)
	}
}

func (t *TableView) addContentRows(statementResults *types.MaterializedStatementResults, truncatedColumnWidths []int) {
	statementResults.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		for colIdx, field := range row.Fields {
			tableCell := tview.NewTableCell(tview.Escape(field.ToString())).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(truncatedColumnWidths[colIdx]).
				SetReference(row)
			t.table.SetCell(rowIdx+1, colIdx, tableCell)
		}
	})
}

func (t *TableView) addPaddingRows() {
	emptyCell := tview.NewTableCell("").
		SetTextColor(tcell.ColorWhite).
		SetAlign(tview.AlignLeft).
		SetSelectable(false)

	for i := 0; i < numPaddingRows; i++ {
		row := t.table.GetRowCount()
		for col := 0; col < t.table.GetColumnCount(); col++ {
			t.table.SetCell(row, col, emptyCell)
		}
	}
}

func (t *TableView) selectLastRow(enableRowSelection bool) {
	t.table.SetSelectable(enableRowSelection, false).Select(t.getLastRowIdx(), 0)
	t.table.ScrollToEnd()
}

func (t *TableView) JumpUp() {
	t.table.Select(t.getSelectedRowIdx()-t.getNumRowsToScroll(), 0)
}

func (t *TableView) getNumRowsToScroll() int {
	_, _, _, numVisibleRows := t.table.GetInnerRect()
	numRowsWithoutHeaderRow := numVisibleRows - 1
	return numRowsWithoutHeaderRow - 1
}

func (t *TableView) JumpDown() {
	t.table.Select(t.getSelectedRowIdx()+t.getNumRowsToScroll(), 0)
}

func (t *TableView) getTableShortcuts(statementResults *types.MaterializedStatementResults, fetchState types.FetchState) []types.Shortcut {
	toggleTableModeText := "Show table"
	if statementResults.IsTableMode() {
		toggleTableModeText = "Show changelog"
	}

	if fetchState == types.Completed {
		return t.getTableShortcutsForCompletedFetchState(toggleTableModeText)
	}

	toggleAutoRefreshText := "Play"
	if fetchState == types.Running {
		toggleAutoRefreshText = "Pause"
	}
	return t.getTableShortcutsForNonCompletedFetchState(toggleTableModeText, toggleAutoRefreshText)
}

func (t *TableView) getTableShortcutsForCompletedFetchState(toggleTableModeText string) []types.Shortcut {
	return []types.Shortcut{
		{KeyText: ExitTableViewShortcut, Text: "Quit"},
		{KeyText: ToggleTableModeShortcut, Text: toggleTableModeText},
		{KeyText: fmt.Sprintf("%s/%s", JumpUpShortcut, JumpDownShortcut), Text: "Jump up/down"},
	}
}

func (t *TableView) getTableShortcutsForNonCompletedFetchState(toggleTableModeText, toggleAutoRefreshText string) []types.Shortcut {
	return []types.Shortcut{
		{KeyText: ExitTableViewShortcut, Text: "Quit"},
		{KeyText: ToggleTableModeShortcut, Text: toggleTableModeText},
		{KeyText: ToggleAutoRefreshShortcut, Text: toggleAutoRefreshText},
		{KeyText: fmt.Sprintf("%s/%s", JumpUpShortcut, JumpDownShortcut), Text: "Jump up/down"},
	}
}
