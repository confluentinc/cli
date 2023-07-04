package controller

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/samber/lo"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type TableController struct {
	table               *tview.Table
	appController       types.ApplicationControllerInterface
	fetchController     types.FetchControllerInterface
	runInteractiveInput func()
	tableLock           sync.Mutex
	isRowViewOpen       bool
	tableWidth          int
	numRowsToScroll     int
	debug               bool
}

func NewTableController(table *tview.Table, appController types.ApplicationControllerInterface, fetchController types.FetchControllerInterface, debug bool) types.TableControllerInterface {
	return &TableController{
		table:           table,
		appController:   appController,
		fetchController: fetchController,
		debug:           debug,
	}
}

func (t *TableController) SetRunInteractiveInputCallback(runInteractiveInput func()) {
	t.runInteractiveInput = runInteractiveInput
}

func (t *TableController) Init(statement types.ProcessedStatement) {
	t.isRowViewOpen = false
	t.fetchController.Init(statement)
	t.fetchController.SetAutoRefreshCallback(t.renderTableAsync)
	t.renderTable()
}

func (t *TableController) renderTableAsync() {
	t.appController.TView().QueueUpdateDraw(t.renderTable)
}

// Function to handle shortcuts and keybindings for TView
func (t *TableController) AppInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if t.isRowViewOpen {
		return t.inputHandlerRowView(event)
	}
	return t.inputHandlerTableView(event)
}

func (t *TableController) inputHandlerRowView(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		char := unicode.ToUpper(event.Rune())
		switch char {
		case 'Q':
			t.closeRowView()
		}
		return nil
	case tcell.KeyCtrlQ:
		fallthrough
	case tcell.KeyEscape:
		t.closeRowView()
		return nil
	}
	return event
}

func (t *TableController) closeRowView() {
	t.appController.ShowTableView()
	t.appController.TView().SetFocus(t.table)
	t.isRowViewOpen = false
}

func (t *TableController) inputHandlerTableView(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		char := unicode.ToUpper(event.Rune())
		action := t.getActionForShortcut(string(char))
		if action != nil {
			action()
		}
		return nil
	case tcell.KeyEscape:
		t.exitTViewMode()
		return nil
	case tcell.KeyCtrlQ:
		t.exitTViewMode()
		return nil
	case tcell.KeyEnter:
		t.openRowView()
		return nil
	}
	return event
}

func (t *TableController) getActionForShortcut(shortcut string) func() {
	switch shortcut {
	case "Q":
		return t.exitTViewMode
	case "M":
		return t.toggleTableModeAndRender
	case "A":
		return t.toggleAutoRefreshAndRender
	case "H":
		return t.fastScrollUp
	case "L":
		return t.fastScrollDown
	}
	return nil
}

func (t *TableController) exitTViewMode() {
	t.fetchController.Close()
	t.appController.SuspendOutputMode(func() {
		output.Println("Result retrieval aborted.")
		t.runInteractiveInput()
	})
}

func (t *TableController) toggleTableModeAndRender() {
	t.fetchController.ToggleTableMode()
	t.renderTable()
}

func (t *TableController) toggleAutoRefreshAndRender() {
	t.fetchController.ToggleAutoRefresh()
	t.renderTable()
}

func (t *TableController) fastScrollUp() {
	currentSelectedRow, _ := t.table.GetSelection()
	rowToSelect := lo.Max([]int{1, currentSelectedRow - t.numRowsToScroll})
	t.table.Select(rowToSelect, 0)
}

func (t *TableController) fastScrollDown() {
	currentSelectedRow, _ := t.table.GetSelection()
	rowToSelect := lo.Min([]int{t.table.GetRowCount() - 1, currentSelectedRow + t.numRowsToScroll})
	t.table.Select(rowToSelect, 0)
}

func (t *TableController) openRowView() {
	if !t.fetchController.IsAutoRefreshRunning() {
		row := t.table.GetCell(t.table.GetSelection()).GetReference().(*types.StatementResultRow)
		t.isRowViewOpen = true

		headers := t.fetchController.GetHeaders()
		sb := strings.Builder{}
		for rowIdx, field := range row.GetFields() {
			sb.WriteString(fmt.Sprintf("[yellow]%s:\n[white]%s\n\n", tview.Escape(headers[rowIdx]), tview.Escape(field.ToString())))
		}
		textView := tview.NewTextView().SetText(sb.String())
		// mouse needs to be disabled, otherwise selecting text with the cursor won't work
		t.appController.TView().SetRoot(components.CreateRowView(textView), true).EnableMouse(false)
		t.appController.TView().SetFocus(textView)
	}
}

func (t *TableController) renderTable() {
	t.tableLock.Lock()
	defer t.tableLock.Unlock()

	t.renderTitle()
	t.renderData()
	t.selectLastRow()
	t.appController.TView().SetFocus(t.table)
}

func (t *TableController) renderTitle() {
	mode := "Changelog mode"
	if t.fetchController.IsTableMode() {
		mode = "Table mode"
	}

	var state string
	switch t.fetchController.GetFetchState() {
	case types.Completed:
		state = "completed"
	case types.Failed:
		state = "auto refresh failed"
	case types.Paused:
		state = "auto refresh paused"
	case types.Running:
		state = fmt.Sprintf("auto refresh %.1fs", float64(defaultRefreshInterval)/1000)
	default:
		state = "unknown error"
	}

	if t.debug {
		t.table.SetTitle(fmt.Sprintf(
			" %s (%s) | last page size: %d | current cache size: %d/%d | table size: %d ",
			mode,
			state,
			t.fetchController.GetStatement().GetPageSize(),
			t.fetchController.GetMaterializedStatementResults().GetChangelogSize(),
			t.fetchController.GetMaterializedStatementResults().GetMaxResults(),
			t.fetchController.GetMaterializedStatementResults().GetTableSize(),
		))
	} else {
		t.table.SetTitle(fmt.Sprintf(" %s (%s) ", mode, state))
	}
}

func (t *TableController) renderData() {
	t.table.Clear()

	_, _, tableWidth, _ := t.table.GetInnerRect()
	t.tableWidth = tableWidth
	columnWidths := t.fetchController.GetMaxWidthPerColumn()
	truncatedColumnWidths := results.GetTruncatedColumnWidths(columnWidths, tableWidth)

	// Print header
	for colIdx, column := range t.fetchController.GetHeaders() {
		tableCell := tview.NewTableCell(column).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false).
			SetMaxWidth(truncatedColumnWidths[colIdx])
		t.table.SetCell(0, colIdx, tableCell)
	}

	// Print content
	t.fetchController.ForEach(t.fillTable(truncatedColumnWidths))

	// add callback function for after draw (gets triggered on any render event, such as screen size update)
	t.table.SetDrawFunc(t.resizeTable(columnWidths))
}

func (t *TableController) fillTable(truncatedColumnWidths []int) func(rowIdx int, row *types.StatementResultRow) {
	return func(rowIdx int, row *types.StatementResultRow) {
		for colIdx, field := range row.Fields {
			tableCell := tview.NewTableCell(tview.Escape(field.ToString())).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(truncatedColumnWidths[colIdx]).
				SetReference(row)
			t.table.SetCell(rowIdx+1, colIdx, tableCell)
		}
	}
}

func (t *TableController) resizeTable(columnWidths []int) func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
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

func (t *TableController) selectLastRow() {
	t.table.SetSelectable(!t.fetchController.IsAutoRefreshRunning(), false).Select(t.table.GetRowCount()-1, 0)
	t.table.ScrollToEnd()
}
