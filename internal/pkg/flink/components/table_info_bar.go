package components

import (
	"fmt"
	"time"

	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type TableInfoBar struct {
	infoBar              *tview.Flex
	refreshState         types.RefreshState
	selectedRowIdx       int
	totalNumRows         int
	lastRefreshTimestamp *time.Time
}

func NewTableInfoBar() *TableInfoBar {
	return &TableInfoBar{
		infoBar: tview.NewFlex().SetDirection(tview.FlexColumn),
	}
}

func (t *TableInfoBar) GetView() *tview.Flex {
	return t.infoBar
}

func (t *TableInfoBar) SetRowInfo(selectedRowIdx, totalNumRows int) {
	t.selectedRowIdx = selectedRowIdx
	t.totalNumRows = totalNumRows
	t.updateInfoBar()
}

func (t *TableInfoBar) SetLastRefreshTimestamp(lastRefreshTimestamp *time.Time) {
	t.lastRefreshTimestamp = lastRefreshTimestamp
	t.updateInfoBar()
}

func (t *TableInfoBar) SetRefreshState(refreshState types.RefreshState) {
	t.refreshState = refreshState
	t.updateInfoBar()
}

func (t *TableInfoBar) updateInfoBar() {
	t.infoBar.Clear()
	t.infoBar.
		AddItem(t.constructRefreshInfo(), 0, 1, false).
		AddItem(t.constructRowInfo(), 0, 1, false).
		AddItem(t.constructLastRefreshInfo(), 0, 1, false)
}

func (t *TableInfoBar) constructRefreshInfo() tview.Primitive {
	return tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(fmt.Sprintf("Refresh: [darkcyan]%s[white]", t.refreshState.ToString()))
}

func (t *TableInfoBar) constructRowInfo() tview.Primitive {
	rowInfo := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter).SetText("")
	if t.selectedRowIdx > 0 && t.totalNumRows > 0 {
		rowInfo.SetText(fmt.Sprintf("Row: [darkcyan]%v[white] of [darkcyan]%v[white]", t.selectedRowIdx, t.totalNumRows))
	}
	return rowInfo
}

func (t *TableInfoBar) constructLastRefreshInfo() tview.Primitive {
	lastRefreshInfo := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignRight).SetText("Last refresh: [darkcyan]-[white]")
	if t.lastRefreshTimestamp != nil {
		lastRefreshInfo.SetText(fmt.Sprintf("Last refresh: [darkcyan]%s[white]", t.lastRefreshTimestamp.Format("15:04:05.000")))
	}
	return lastRefreshInfo
}
