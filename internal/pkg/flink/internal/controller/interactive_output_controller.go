package controller

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type InteractiveOutputController struct {
	app           *tview.Application
	tableView     components.TableViewInterface
	resultFetcher types.ResultFetcherInterface
	isRowViewOpen bool
	debug         bool
}

func NewInteractiveOutputController(tableView components.TableViewInterface, resultFetcher types.ResultFetcherInterface, debug bool) types.OutputControllerInterface {
	return &InteractiveOutputController{
		app:           tview.NewApplication(),
		tableView:     tableView,
		resultFetcher: resultFetcher,
		debug:         debug,
	}
}

func (t *InteractiveOutputController) VisualizeResults() {
	t.init()
	t.start()
}

func (t *InteractiveOutputController) start() {
	err := t.app.Run()
	if err != nil {
		log.CliLogger.Errorf("Error: failed to open table view, %v", err)
		utils.OutputErr("Error: failed to open table view")
	}
}

func (t *InteractiveOutputController) init() {
	t.isRowViewOpen = false
	t.resultFetcher.SetRefreshCallback(t.renderTableAsync)
	t.resultFetcher.ToggleRefresh()
	t.app.SetInputCapture(t.inputCapture)
	t.tableView.Init(t.app.QueueUpdateDraw)
	t.updateTable()
}

func (t *InteractiveOutputController) updateTable() {
	t.tableView.RenderTable(t.getTableTitle(), t.resultFetcher.GetMaterializedStatementResults(), t.resultFetcher.GetLastRefreshTimestamp(), t.resultFetcher.GetRefreshState())
	t.renderTableView()
}

func (t *InteractiveOutputController) renderTableView() {
	t.app.SetRoot(t.tableView.GetRoot(), true).EnableMouse(false)
	t.app.SetFocus(t.tableView.GetFocusableElement())
}

func (t *InteractiveOutputController) renderTableAsync() {
	t.app.QueueUpdateDraw(t.updateTable)
}

// Function to handle shortcuts and keybindings for TView
func (t *InteractiveOutputController) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	if t.isRowViewOpen {
		return t.inputHandlerRowView(event)
	}
	return t.inputHandlerTableView(event)
}

func (t *InteractiveOutputController) inputHandlerRowView(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		shortcut := string(unicode.ToUpper(event.Rune()))
		switch shortcut {
		case components.ExitRowViewShortcut:
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

func (t *InteractiveOutputController) closeRowView() {
	t.renderTableView()
	t.isRowViewOpen = false
}

func (t *InteractiveOutputController) inputHandlerTableView(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		char := unicode.ToUpper(event.Rune())
		action := t.getActionForShortcut(string(char))
		if action != nil {
			action()
		}
		return nil
	case tcell.KeyEscape:
		fallthrough
	case tcell.KeyCtrlQ:
		t.exitTViewMode()
		return nil
	case tcell.KeyEnter:
		t.renderRowView()
		return nil
	case tcell.KeyUp:
		fallthrough
	case tcell.KeyDown:
		return t.handleKeyUpOrDownPress(event)
	}
	return event
}

func (t *InteractiveOutputController) getActionForShortcut(shortcut string) func() {
	switch shortcut {
	case components.ExitTableViewShortcut:
		return t.exitTViewMode
	case components.ToggleTableModeShortcut:
		return t.toggleTableMode
	case components.ToggleRefreshShortcut:
		return t.toggleRefresh
	case components.JumpUpShortcut:
		return t.stopRefreshOrScroll(t.tableView.JumpUp)
	case components.JumpDownShortcut:
		return t.stopRefreshOrScroll(t.tableView.JumpDown)
	}
	return nil
}

func (t *InteractiveOutputController) exitTViewMode() {
	t.resultFetcher.Close()
	t.app.Stop()
	output.Println("Result retrieval aborted.")
}

func (t *InteractiveOutputController) toggleTableMode() {
	t.resultFetcher.ToggleTableMode()
	t.updateTable()
}

func (t *InteractiveOutputController) toggleRefresh() {
	if t.resultFetcher.GetRefreshState() != types.Completed {
		t.resultFetcher.ToggleRefresh()
		t.updateTable()
	}
}

func (t *InteractiveOutputController) stopRefreshOrScroll(scroll func()) func() {
	if t.resultFetcher.IsRefreshRunning() {
		return func() {
			t.resultFetcher.ToggleRefresh()
			t.updateTable()
		}
	}
	return scroll
}

func (t *InteractiveOutputController) renderRowView() {
	if !t.resultFetcher.IsRefreshRunning() {
		row := t.tableView.GetSelectedRow()
		t.isRowViewOpen = true

		headers := t.resultFetcher.GetMaterializedStatementResults().GetHeaders()
		sb := strings.Builder{}
		for rowIdx, field := range row.GetFields() {
			sb.WriteString(fmt.Sprintf("[yellow]%s:\n[white]%s\n\n", tview.Escape(headers[rowIdx]), tview.Escape(field.ToString())))
		}
		textView := tview.NewTextView().SetText(sb.String())
		// mouse needs to be disabled, otherwise selecting text with the cursor won't work
		t.app.SetRoot(components.CreateRowView(textView), true).EnableMouse(false)
		t.app.SetFocus(textView)
	}
}

func (t *InteractiveOutputController) handleKeyUpOrDownPress(event *tcell.EventKey) *tcell.EventKey {
	if t.resultFetcher.IsRefreshRunning() {
		t.resultFetcher.ToggleRefresh()
		t.updateTable()
		return nil
	}
	return event
}

func (t *InteractiveOutputController) getTableTitle() string {
	mode := "Changelog mode"
	if t.resultFetcher.IsTableMode() {
		mode = "Table mode"
	}

	if t.debug {
		return fmt.Sprintf(
			" %s | last page size: %d | current cache size: %d/%d | table size: %d ",
			mode,
			t.resultFetcher.GetStatement().GetPageSize(),
			t.resultFetcher.GetMaterializedStatementResults().GetChangelogSize(),
			t.resultFetcher.GetMaterializedStatementResults().GetMaxResults(),
			t.resultFetcher.GetMaterializedStatementResults().GetTableSize(),
		)
	}

	return fmt.Sprintf(" %s ", mode)
}
