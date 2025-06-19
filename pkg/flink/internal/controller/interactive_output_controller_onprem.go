package controller

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/confluentinc/cli/v4/pkg/flink/components"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type InteractiveOutputControllerOnPrem struct {
	app            *tview.Application
	tableView      components.TableViewInterface
	resultFetcher  types.ResultFetcherInterfaceOnPrem
	isRowViewOpen  bool
	userProperties types.UserPropertiesInterface
	debug          bool
}

func NewInteractiveOutputControllerOnPrem(tableView components.TableViewInterface, resultFetcher types.ResultFetcherInterfaceOnPrem, userProperties types.UserPropertiesInterface, debug bool) types.OutputControllerInterface {
	return &InteractiveOutputControllerOnPrem{
		app:            tview.NewApplication(),
		tableView:      tableView,
		resultFetcher:  resultFetcher,
		userProperties: userProperties,
		debug:          debug,
	}
}

func (t *InteractiveOutputControllerOnPrem) VisualizeResults() {
	t.init()
	t.startTView() // this is blocking
	t.close()
}

func (t *InteractiveOutputControllerOnPrem) startTView() {
	err := t.app.Run()
	if err != nil {
		log.CliLogger.Errorf("%s, %v", errorDuringTableView, err)
		utils.OutputErr(errorDuringTableView)
	}
}

func (t *InteractiveOutputControllerOnPrem) close() {
	t.resultFetcher.Close()
	output.Println(false, "Result retrieval aborted.")
}

func (t *InteractiveOutputControllerOnPrem) init() {
	t.isRowViewOpen = false
	t.resultFetcher.SetRefreshCallback(t.renderTableAsync)
	t.resultFetcher.ToggleRefresh()
	t.app.SetInputCapture(t.inputCapture)
	t.tableView.Init()
	t.updateTable()
}

func (t *InteractiveOutputControllerOnPrem) updateTable() {
	t.tableView.GetFocusableElement().SetBorder(t.withBorder())
	t.tableView.RenderTable(t.getTableTitle(), t.resultFetcher.GetMaterializedStatementResults(), t.resultFetcher.GetLastRefreshTimestamp(), t.resultFetcher.GetRefreshState())
	t.renderTableView()
}

func (t *InteractiveOutputControllerOnPrem) renderTableView() {
	t.app.SetRoot(t.tableView.GetRoot(), true).EnableMouse(false)
	t.app.SetFocus(t.tableView.GetFocusableElement())
}

func (t *InteractiveOutputControllerOnPrem) renderTableAsync() {
	t.app.QueueUpdateDraw(t.updateTable)
}

// Function to handle shortcuts and keybindings for TView
func (t *InteractiveOutputControllerOnPrem) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	if t.isRowViewOpen {
		return t.inputHandlerRowView(event)
	}
	return t.inputHandlerTableView(event)
}

func (t *InteractiveOutputControllerOnPrem) inputHandlerRowView(event *tcell.EventKey) *tcell.EventKey {
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

func (t *InteractiveOutputControllerOnPrem) closeRowView() {
	t.renderTableView()
	t.isRowViewOpen = false
}

func (t *InteractiveOutputControllerOnPrem) inputHandlerTableView(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyRune:
		char := unicode.ToUpper(event.Rune())
		action := t.getActionForShortcut(string(char))
		if action != nil {
			action()
		}
		return nil
	case tcell.KeyEscape, tcell.KeyCtrlQ:
		t.app.Stop()
		return nil
	case tcell.KeyEnter:
		t.renderRowView()
		return nil
	case tcell.KeyUp, tcell.KeyDown:
		return t.handleKeyUpOrDownPress(event)
	}
	return event
}

func (t *InteractiveOutputControllerOnPrem) getActionForShortcut(shortcut string) func() {
	switch shortcut {
	case components.ExitTableViewShortcut:
		return t.app.Stop
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

func (t *InteractiveOutputControllerOnPrem) toggleTableMode() {
	t.resultFetcher.ToggleTableMode()
	t.updateTable()
}

func (t *InteractiveOutputControllerOnPrem) toggleRefresh() {
	if t.resultFetcher.GetRefreshState() != types.Completed {
		t.resultFetcher.ToggleRefresh()
		t.updateTable()
	}
}

func (t *InteractiveOutputControllerOnPrem) stopRefreshOrScroll(scroll func()) func() {
	if t.resultFetcher.IsRefreshRunning() {
		return func() {
			t.resultFetcher.ToggleRefresh()
			t.updateTable()
		}
	}
	return scroll
}

func (t *InteractiveOutputControllerOnPrem) renderRowView() {
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
		rowView := components.CreateRowView(textView, t.withBorder())
		t.app.SetRoot(rowView, true).EnableMouse(false)
		t.app.SetFocus(textView)
	}
}

func (t *InteractiveOutputControllerOnPrem) handleKeyUpOrDownPress(event *tcell.EventKey) *tcell.EventKey {
	if t.resultFetcher.IsRefreshRunning() {
		t.resultFetcher.ToggleRefresh()
		t.updateTable()
		return nil
	}
	return event
}

func (t *InteractiveOutputControllerOnPrem) withBorder() bool {
	return t.userProperties.GetOutputFormat() != config.OutputFormatPlainText
}

func (t *InteractiveOutputControllerOnPrem) getTableTitle() string {
	mode := "Changelog mode"
	if t.resultFetcher.IsTableMode() {
		mode = "Table mode"
	}

	if t.debug {
		return fmt.Sprintf(
			" %s (%s) | last page size: %d | current cache size: %d/%d | table size: %d ",
			mode,
			t.resultFetcher.GetStatement().StatementName,
			t.resultFetcher.GetStatement().GetPageSize(),
			t.resultFetcher.GetMaterializedStatementResults().GetChangelogSize(),
			t.resultFetcher.GetMaterializedStatementResults().GetMaxResults(),
			t.resultFetcher.GetMaterializedStatementResults().GetTableSize(),
		)
	}

	return fmt.Sprintf(" %s (%s) ", mode, t.resultFetcher.GetStatement().StatementName)
}
