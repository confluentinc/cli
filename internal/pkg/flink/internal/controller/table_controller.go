package controller

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type TableController struct {
	app           *tview.Application
	tableView     *components.TableView
	resultFetcher types.ResultFetcherInterface
	isRowViewOpen bool
}

func NewTableController(resultFetcher types.ResultFetcherInterface) types.OutputControllerInterface {
	return &TableController{
		app:           tview.NewApplication(),
		resultFetcher: resultFetcher,
	}
}

func (t *TableController) VisualizeResults() {
	t.Init(t.processedStatement)
	t.Start()
}

func (t *TableController) Start() {
	err := t.app.Run()
	if err != nil {
		log.CliLogger.Errorf("Failed to start tview., %v", err)
		utils.OutputErr("Error: failed to open table view")
	}
}

func (t *TableController) Init(statement types.ProcessedStatement) {
	t.isRowViewOpen = false
	t.resultFetcher.Init(statement)
	t.resultFetcher.SetAutoRefreshCallback(t.renderTableAsync)
	t.app.SetInputCapture(t.inputCapture)
	t.initTableView()
}

func (t *TableController) initTableView() {
	t.tableView = components.NewTableView()
	t.updateTable()
	t.openTableView()
}

func (t *TableController) updateTable() {
	t.tableView.RenderTable(t.getTableTitle(), t.resultFetcher.GetResults(), !t.resultFetcher.IsAutoRefreshRunning())
	t.app.SetFocus(t.tableView.GetTable())
}

func (t *TableController) openTableView() {
	t.app.SetRoot(t.tableView.RootLayout, true).EnableMouse(false)
	t.app.SetFocus(t.tableView.GetTable())
}

func (t *TableController) renderTableAsync() {
	t.app.QueueUpdateDraw(t.updateTable)
}

// Function to handle shortcuts and keybindings for TView
func (t *TableController) inputCapture(event *tcell.EventKey) *tcell.EventKey {
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
	t.openTableView()
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
		return t.renderAfterAction(t.resultFetcher.ToggleTableMode)
	case "A":
		return t.renderAfterAction(t.resultFetcher.ToggleAutoRefresh)
	case "N":
		return t.renderAfterAction(t.resultFetcher.FetchNextPage)
	case "R":
		return t.renderAfterAction(t.resultFetcher.JumpToLastPage)
	case "H":
		return t.tableView.FastScrollUp
	case "L":
		return t.tableView.FastScrollDown
	}
	return nil
}

func (t *TableController) exitTViewMode() {
	t.resultFetcher.Close()
	t.app.Stop()
	output.Println("Result retrieval aborted.")
}

func (t *TableController) renderAfterAction(action func()) func() {
	return func() {
		action()
		t.updateTable()
	}
}

func (t *TableController) openRowView() {
	if !t.resultFetcher.IsAutoRefreshRunning() {
		row := t.tableView.GetSelectedRow()
		t.isRowViewOpen = true

		headers := t.resultFetcher.GetResults().GetHeaders()
		sb := strings.Builder{}
		for rowIdx, field := range row.Fields {
			sb.WriteString(fmt.Sprintf("[yellow]%s:\n[white]%s\n\n", tview.Escape(headers[rowIdx]), tview.Escape(field.ToString())))
		}
		textView := tview.NewTextView().SetText(sb.String())
		// mouse needs to be disabled, otherwise selecting text with the cursor won't work
		t.app.SetRoot(components.CreateRowView(textView), true).EnableMouse(false)
		t.app.SetFocus(textView)
	}
}

func (t *TableController) getTableTitle() string {
	mode := "Changelog mode"
	if t.resultFetcher.IsTableMode() {
		mode = "Table mode"
	}

	var state string
	switch t.resultFetcher.GetFetchState() {
	case types.Completed:
		state = "completed"
	case types.Failed:
		state = "auto refresh failed"
	case types.Paused:
		state = "auto refresh paused"
	case types.Running:
		state = fmt.Sprintf("auto refresh %vs", results.DefaultRefreshInterval/1000)
	default:
		state = "unknown error"
	}

	return fmt.Sprintf(" %s (%s) ", mode, state)
}
