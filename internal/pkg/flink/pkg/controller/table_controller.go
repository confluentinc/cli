package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/confluentinc/flink-sql-client/pkg/results"

	"github.com/confluentinc/flink-sql-client/pkg/types"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type TableControllerInterface interface {
	AppInputCapture(event *tcell.EventKey) *tcell.EventKey
	Init(statement types.ProcessedStatement)
	SetRunInteractiveInputCallback(func())
	GetActionForShortcut(shortcut string) func()
}

type TableController struct {
	table                        *tview.Table
	appController                ApplicationControllerInterface
	runInteractiveInput          func()
	store                        StoreInterface
	statement                    types.ProcessedStatement
	materializedStatementResults results.MaterializedStatementResults
	hasUserDisabledTableMode     bool
	hasUserDisabledAutoFetch     bool
	cancelFetch                  context.CancelFunc
	tableLock                    sync.Mutex
	cancelLock                   sync.RWMutex
}

const maxResultsCapacity int = 1000
const defaultRefreshInterval uint = 1000 // in milliseconds
const minRefreshInterval uint = 100      // in milliseconds

func NewTableController(tableRef *tview.Table, store StoreInterface, appController ApplicationControllerInterface) TableControllerInterface {
	controller := &TableController{
		table:         tableRef,
		appController: appController,
		store:         store,
	}
	return controller
}

func (t *TableController) SetRunInteractiveInputCallback(runInteractiveInput func()) {
	t.runInteractiveInput = runInteractiveInput
}

func (t *TableController) exitTViewMode() {
	t.stopAutoRefresh()
	t.store.DeleteStatement(t.statement.StatementName)
	t.appController.SuspendOutputMode(t.runInteractiveInput)
}

func (t *TableController) GetActionForShortcut(shortcut string) func() {
	switch shortcut {
	case "Q":
		return func() {
			t.exitTViewMode()
		}
	case "M":
		return func() {
			t.hasUserDisabledTableMode = !t.hasUserDisabledTableMode
			t.materializedStatementResults.SetTableMode(!t.hasUserDisabledTableMode)
			t.renderTable()
		}
	case "R":
		return func() {
			if t.hasUserDisabledAutoFetch {
				t.hasUserDisabledAutoFetch = false
				t.startAutoRefresh(t.statement, defaultRefreshInterval)
			} else {
				t.hasUserDisabledAutoFetch = true
				t.stopAutoRefresh()
			}
			t.renderTable()
		}
	}
	return nil
}

// Function to handle shortcuts and keybindings for TView
func (t *TableController) AppInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		char := unicode.ToUpper(event.Rune())
		action := t.GetActionForShortcut(string(char))
		if action != nil {
			action()
		}
		return nil
	} else {
		switch event.Key() {
		case tcell.KeyCtrlC:
			t.onCtrlC()
			return nil
		case tcell.KeyEscape:
			t.exitTViewMode()
			return nil
		case tcell.KeyCtrlQ:
			t.exitTViewMode()
			return nil
		}
	}
	return event
}

func (t *TableController) setRefreshCancelFunc(cancelFunc context.CancelFunc) {
	t.cancelLock.Lock()
	defer t.cancelLock.Unlock()

	t.cancelFetch = cancelFunc
}

func (t *TableController) stopAutoRefresh() {
	if t.isAutoRefreshRunning() {
		t.cancelFetch()
		t.setRefreshCancelFunc(nil)
	}
}

func (t *TableController) startAutoRefresh(statement types.ProcessedStatement, refreshInterval uint) {
	if statement.PageToken == "" || t.isAutoRefreshRunning() {
		return
	}
	fetchCtx, cancelFetch := context.WithCancel(context.Background())
	t.setRefreshCancelFunc(cancelFetch)
	t.refreshResults(fetchCtx, statement, refreshInterval)
}

func (t *TableController) isAutoRefreshRunning() bool {
	t.cancelLock.RLock()
	defer t.cancelLock.RUnlock()

	return t.cancelFetch != nil
}

func (t *TableController) refreshResults(ctx context.Context, statement types.ProcessedStatement, refreshInterval uint) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				t.renderTable()
				t.appController.TView().Draw()

				newResults, err := t.store.FetchStatementResults(statement)
				if err != nil {
					continue
				}

				// don't fetch if we have a next page token or the refresh interval is < min
				if newResults.PageToken == "" || refreshInterval < minRefreshInterval {
					t.stopAutoRefresh()
					continue
				}

				statement = *newResults
				t.materializedStatementResults.AppendAll(newResults.StatementResults.GetRows())
				time.Sleep(time.Millisecond * time.Duration(refreshInterval))
			}
		}
	}()
}

func (t *TableController) Init(statement types.ProcessedStatement) {
	t.statement = statement
	t.materializedStatementResults = results.NewMaterializedStatementResults(statement.StatementResults.GetHeaders(), maxResultsCapacity)
	t.materializedStatementResults.SetTableMode(!t.hasUserDisabledTableMode)
	t.materializedStatementResults.AppendAll(statement.StatementResults.GetRows())
	// if unbounded result start refreshing results in the background
	if statement.PageToken != "" && !t.hasUserDisabledAutoFetch {
		t.startAutoRefresh(statement, defaultRefreshInterval)
	} else {
		t.renderTable()
	}
}

func (t *TableController) renderTable() {
	t.tableLock.Lock()
	defer t.tableLock.Unlock()

	t.renderTitle()
	t.renderData()
	t.focus()
}

func (t *TableController) renderTitle() {
	mode := "Changelog mode"
	if t.materializedStatementResults.IsTableMode() {
		mode = "Table mode"
	}

	if t.statement.PageToken == "" {
		t.table.SetTitle(fmt.Sprintf("%s (completed)", mode))
	} else {
		if t.isAutoRefreshRunning() {
			t.table.SetTitle(fmt.Sprintf("%s (auto refresh %vs)", mode, defaultRefreshInterval/1000))
		} else {
			t.table.SetTitle(fmt.Sprintf("%s (auto refresh disabled)", mode))
		}
	}
}

func (t *TableController) renderData() {
	t.table.Clear()
	// Print header
	for colIdx, column := range t.materializedStatementResults.GetHeaders() {
		tableCell := tview.NewTableCell(column).
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignLeft).
			SetSelectable(false)
		t.table.SetCell(0, colIdx, tableCell)
	}

	rowIdx := 1
	iterator := t.materializedStatementResults.Iterator()
	// Print content
	for iterator.HasNext() {
		row := iterator.GetNext()
		for colIdx, field := range row.Fields {
			color := tcell.ColorWhite
			tableCell := tview.NewTableCell(tview.Escape(field.Format(nil))).
				SetTextColor(color).
				SetAlign(tview.AlignLeft)
			t.table.SetCell(rowIdx, colIdx, tableCell)
		}
		rowIdx++
	}
}

func (t *TableController) selectRow() {
	t.table.SetBorders(false).
		SetSelectable(!t.isAutoRefreshRunning(), false).
		SetSeparator(' ').
		Select(t.table.GetRowCount()-1, 0)
	t.table.ScrollToEnd()
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
