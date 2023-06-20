package controller

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/samber/lo"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/store"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type TableController struct {
	table                                *tview.Table
	appController                        types.ApplicationControllerInterface
	runInteractiveInput                  func()
	store                                store.StoreInterface
	statement                            types.ProcessedStatement
	statementLock                        sync.RWMutex
	materializedStatementResults         results.MaterializedStatementResults
	materializedStatementResultsIterator results.MaterializedStatementResultsIterator
	selectedRowIdx                       int
	tableLock                            sync.Mutex
	isRowViewOpen                        bool
	tableWidth                           int
	fetchState                           int32
}

type fetchState int32

const (
	paused    fetchState = iota // auto fetch was paused
	completed                   // arrived at last page, fetch is completed
	failed                      // fetching next page failed
	running                     // auto fetch is running

	maxResultsCapacity     int  = 1000
	defaultRefreshInterval uint = 1000 // in milliseconds
	minColumnWidth         int  = 4    // min characters displayed in a column
)

func NewTableController(table *tview.Table, store store.StoreInterface, appController types.ApplicationControllerInterface) types.TableControllerInterface {
	return &TableController{
		table:         table,
		appController: appController,
		store:         store,
	}
}

// Function to handle shortcuts and keybindings for TView
func (t *TableController) AppInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if t.isRowViewOpen {
		return t.inputHandlerRowView(event)
	}
	return t.inputHandlerTableView(event)
}

func (t *TableController) Init(statement types.ProcessedStatement) {
	t.setStatement(statement)
	t.materializedStatementResults = results.NewMaterializedStatementResults(statement.StatementResults.GetHeaders(), maxResultsCapacity)
	t.materializedStatementResults.SetTableMode(true)
	t.materializedStatementResults.Append(statement.StatementResults.GetRows()...)
	// if unbounded result start refreshing results in the background
	if statement.PageToken != "" {
		t.startAutoRefresh(defaultRefreshInterval)
	} else {
		t.setFetchState(completed)
		t.renderTable()
	}
}

func (t *TableController) SetRunInteractiveInputCallback(runInteractiveInput func()) {
	t.runInteractiveInput = runInteractiveInput
}

func (t *TableController) exitTViewMode() {
	t.setFetchState(paused)
	// This was used to delete statements after their execution to save system resources, which should not be
	// an issue anymore. We don't want to remove it completely just yet, but will disable it by default for now.
	// TODO: remove this completely once we are sure we won't need it in the future
	statement := t.getStatement()
	if config.ShouldCleanupStatements || statement.Status == types.RUNNING {
		go t.store.DeleteStatement(statement.StatementName)
	}
	t.appController.SuspendOutputMode(func() {
		output.Println("Result retrieval aborted.")
		t.runInteractiveInput()
	})
}

func (t *TableController) toggleTableModeAndRender() {
	t.materializedStatementResults.SetTableMode(!t.materializedStatementResults.IsTableMode())
	t.renderTable()
}

func (t *TableController) toggleAutoRefreshAndRender() {
	if t.isAutoRefreshRunning() {
		t.setFetchState(paused)
		t.renderTable()
		return
	}

	t.startAutoRefresh(defaultRefreshInterval)
}

func (t *TableController) fetchNextPageAndRender() {
	t.fetchNextPage()
	t.renderTable()
}

func (t *TableController) hasMoreResults() bool {
	return len(t.getStatement().StatementResults.GetRows()) > 0 && t.getFetchState() != failed && t.getFetchState() != completed
}

func (t *TableController) goToLastPageAndRender() {
	go func() {
		// fetch next pages until we receive no results (which means we are at the last page)
		for {
			t.fetchNextPage()
			if !t.hasMoreResults() {
				break
			}
			// minimal wait to avoid rate limiting
			time.Sleep(time.Millisecond * 50)
		}
		t.appController.TView().QueueUpdateDraw(t.renderTable)
	}()
}

func (t *TableController) getActionForShortcut(shortcut string) func() {
	switch shortcut {
	case "Q":
		return t.exitTViewMode
	case "M":
		return t.toggleTableModeAndRender
	case "R":
		return t.toggleAutoRefreshAndRender
	case "N":
		return t.fetchNextPageAndRender
	case "L":
		return t.goToLastPageAndRender
	}
	return nil
}

func (t *TableController) openRowView() {
	if !t.isAutoRefreshRunning() {
		row := t.materializedStatementResultsIterator.Value()
		t.isRowViewOpen = true

		headers := t.materializedStatementResults.GetHeaders()
		sb := strings.Builder{}
		for rowIdx, field := range row.Fields {
			sb.WriteString(fmt.Sprintf("[yellow]%s:\n[white]%s\n\n", tview.Escape(headers[rowIdx]), tview.Escape(field.ToString())))
		}
		textView := tview.NewTextView().SetText(sb.String())
		// mouse needs to be disabled, otherwise selecting text with the cursor won't work
		t.appController.TView().SetRoot(components.CreateRowView(textView), true).EnableMouse(false)
		t.appController.TView().SetFocus(textView)
	}
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

func (t *TableController) isAutoRefreshRunning() bool {
	return t.getFetchState() == running
}

func (t *TableController) isAutoRefreshStartAllowed() bool {
	return t.getFetchState() == paused || t.getFetchState() == failed
}

func (t *TableController) fetchNextPage() {
	// don't fetch if we're already at the last page, otherwise we would fetch the first page again
	if t.getFetchState() == completed {
		return
	}

	// fetch
	newResults, err := t.store.FetchStatementResults(t.getStatement())
	if err != nil {
		t.setFetchState(failed)
		return
	}

	// update data
	t.setStatement(*newResults)
	t.materializedStatementResults.Append(newResults.StatementResults.GetRows()...)
	if newResults.PageToken == "" {
		t.setFetchState(completed)
		return
	}

	// if auto refresh is not running we set the state to paused
	if !t.isAutoRefreshRunning() {
		t.setFetchState(paused)
		return
	}

	t.setFetchState(running)
}

func (t *TableController) startAutoRefresh(refreshInterval uint) {
	if t.isAutoRefreshStartAllowed() {
		t.setFetchState(running)
		go func() {
			for t.isAutoRefreshRunning() {
				t.fetchNextPage()
				t.appController.TView().QueueUpdateDraw(t.renderTable)
				time.Sleep(time.Millisecond * time.Duration(refreshInterval))
			}
		}()
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
	if t.materializedStatementResults.IsTableMode() {
		mode = "Table mode"
	}

	var state string
	switch t.getFetchState() {
	case completed:
		state = "completed"
	case failed:
		state = "auto refresh failed"
	case paused:
		state = "auto refresh paused"
	case running:
		state = fmt.Sprintf("auto refresh %vs", defaultRefreshInterval/1000)
	default:
		state = "unknown error"
	}

	t.table.SetTitle(fmt.Sprintf(" %s (%s) ", mode, state))
}

func (t *TableController) rowSelectionHandler(row, col int) {
	// table title (-1) and header row (0) are not selectable
	if row <= 0 {
		row = 1
	}
	// check if selected row is out of bounds
	if row >= t.table.GetRowCount() {
		row = t.table.GetRowCount() - 1
	}
	if !t.isAutoRefreshRunning() {
		stepsToMove := row - t.selectedRowIdx
		t.materializedStatementResultsIterator.Move(stepsToMove)
	}
	t.selectedRowIdx = row
}

func (t *TableController) renderData() {
	t.table.Clear()
	t.table.SetSelectionChangedFunc(t.rowSelectionHandler)

	_, _, tableWidth, _ := t.table.GetInnerRect()
	t.tableWidth = tableWidth
	columnWidths := t.materializedStatementResults.GetMaxWidthPerColum()
	truncatedColumnWidths := results.GetTruncatedColumnWidths(columnWidths, tableWidth)

	// Print header
	for colIdx, column := range t.materializedStatementResults.GetHeaders() {
		tableCell := tview.NewTableCell(column).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false).
			SetMaxWidth(truncatedColumnWidths[colIdx])
		t.table.SetCell(0, colIdx, tableCell)
	}

	// Print content
	t.materializedStatementResults.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		for colIdx, field := range row.Fields {
			tableCell := tview.NewTableCell(tview.Escape(field.ToString())).
				SetTextColor(tcell.ColorWhite).
				SetAlign(tview.AlignLeft).
				SetMaxWidth(truncatedColumnWidths[colIdx])
			t.table.SetCell(rowIdx+1, colIdx, tableCell)
		}
	})

	// add callback function for after draw (gets triggered on any render event, such as screen size update)
	t.table.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
		// check if the table width has changed
		newX, newY, newWidth, newHeight := t.table.GetInnerRect()
		hasTableWidthChanged := t.tableWidth != newWidth
		t.tableWidth = newWidth
		if !hasTableWidthChanged {
			return newX, newY, newWidth, newHeight
		}

		// check if space needed fits screen, if it doesn't truncate the column
		truncatedColumnWidths = results.GetTruncatedColumnWidths(columnWidths, newWidth)
		for rowIdx := 0; rowIdx < t.table.GetRowCount(); rowIdx++ {
			for colIdx := 0; colIdx < t.table.GetColumnCount(); colIdx++ {
				t.table.GetCell(rowIdx, colIdx).SetMaxWidth(lo.Max([]int{truncatedColumnWidths[colIdx], minColumnWidth}))
			}
		}
		return newX, newY, newWidth, newHeight
	})
}

func (t *TableController) selectLastRow() {
	if !t.isAutoRefreshRunning() {
		t.materializedStatementResultsIterator = t.materializedStatementResults.Iterator(true)
		t.selectedRowIdx = t.table.GetRowCount() - 1
	}

	t.table.SetSelectable(!t.isAutoRefreshRunning(), false).
		Select(t.table.GetRowCount()-1, 0)
	t.table.ScrollToEnd()
}

func (t *TableController) getStatement() types.ProcessedStatement {
	t.statementLock.RLock()
	defer t.statementLock.RUnlock()

	return t.statement
}

func (t *TableController) setStatement(statement types.ProcessedStatement) {
	t.statementLock.Lock()
	defer t.statementLock.Unlock()

	t.statement = statement
}

func (t *TableController) setFetchState(state fetchState) {
	atomic.StoreInt32(&t.fetchState, int32(state))
}

func (t *TableController) getFetchState() fetchState {
	return fetchState(atomic.LoadInt32(&t.fetchState))
}
