package results

import (
	"sync"
	"time"

	"github.com/confluentinc/cli/v4/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

type ResultFetcherOnPrem struct {
	store                        types.StoreInterface
	statement                    types.ProcessedStatement
	statementLock                sync.RWMutex
	materializedStatementResults types.MaterializedStatementResults
	refreshState                 refreshState
	refreshCallback              func()
	fetchLock                    sync.Mutex
}

func NewResultFetcherOnPrem(store types.StoreInterface) types.ResultFetcherInterfaceOnPrem {
	return &ResultFetcherOnPrem{
		store:           store,
		refreshCallback: func() {},
	}
}

func (t *ResultFetcherOnPrem) IsTableMode() bool {
	return t.materializedStatementResults.IsTableMode()
}

func (t *ResultFetcherOnPrem) ToggleTableMode() {
	t.materializedStatementResults.SetTableMode(!t.materializedStatementResults.IsTableMode())
}

func (t *ResultFetcherOnPrem) ToggleRefresh() {
	if t.IsRefreshRunning() {
		t.refreshState.setState(types.Paused)
		return
	}

	t.startRefresh(DefaultRefreshInterval)
}

func (t *ResultFetcherOnPrem) IsRefreshRunning() bool {
	return t.GetRefreshState() == types.Running
}

func (t *ResultFetcherOnPrem) GetRefreshState() types.RefreshState {
	return t.refreshState.getState()
}

func (t *ResultFetcherOnPrem) startRefresh(refreshInterval uint) {
	if t.isRefreshStartAllowed() {
		t.refreshState.setState(types.Running)
		go utils.WithPanicRecovery(func() {
			for t.IsRefreshRunning() {
				t.fetchNextPageAndUpdateState()
				// break here to avoid rendering and messing with the view if pause was initiated
				if t.GetRefreshState() == types.Paused {
					break
				}
				t.refreshCallback()
				time.Sleep(time.Millisecond * time.Duration(refreshInterval))
			}
		})()
	}
}

func (t *ResultFetcherOnPrem) isRefreshStartAllowed() bool {
	return t.GetRefreshState() == types.Paused || t.GetRefreshState() == types.Failed
}

func (t *ResultFetcherOnPrem) fetchNextPageAndUpdateState() {
	// lock here to make sure we don't fetch the same page twice
	t.fetchLock.Lock()
	defer t.fetchLock.Unlock()

	newResults, err := t.store.FetchStatementResults(t.GetStatement())
	t.updateState(newResults, err)
}

func (t *ResultFetcherOnPrem) updateState(newResults *types.ProcessedStatement, err *types.StatementError) {
	// don't fetch if we're already at the last page, otherwise we would fetch the first page again
	if t.GetRefreshState() == types.Completed {
		return
	}

	if err != nil {
		t.refreshState.setState(types.Failed)
		return
	}

	t.setStatement(*newResults)
	t.materializedStatementResults.Append(newResults.StatementResults.GetRows()...)
	if newResults.PageToken == "" {
		t.refreshState.setState(types.Completed)
		return
	}

	// if auto refresh is not running we set the state to types.Paused
	if !t.IsRefreshRunning() {
		t.refreshState.setState(types.Paused)
		return
	}

	t.refreshState.setState(types.Running)
}

func (t *ResultFetcherOnPrem) GetStatement() types.ProcessedStatement {
	t.statementLock.RLock()
	defer t.statementLock.RUnlock()

	return t.statement
}

func (t *ResultFetcherOnPrem) setStatement(statement types.ProcessedStatement) {
	t.statementLock.Lock()
	defer t.statementLock.Unlock()

	t.statement = statement
}

func (t *ResultFetcherOnPrem) Init(statement types.ProcessedStatement) {
	t.setStatement(statement)
	t.setInitialRefreshState(statement)
	headers := t.getResultHeadersOrCreateFromResultSchema(statement)
	t.materializedStatementResults = types.NewMaterializedStatementResults(headers, MaxResultsCapacity, statement.Traits.GetUpsertColumns())
	t.materializedStatementResults.SetTableMode(true)
	t.materializedStatementResults.Append(statement.StatementResults.GetRows()...)
}

func (t *ResultFetcherOnPrem) setInitialRefreshState(statement types.ProcessedStatement) {
	if statement.PageToken == "" {
		t.refreshState.setState(types.Completed)
		return
	}
	t.refreshState.setState(types.Paused)
}

func (t *ResultFetcherOnPrem) getResultHeadersOrCreateFromResultSchema(statement types.ProcessedStatement) []string {
	if len(statement.StatementResults.GetHeaders()) > 0 {
		return statement.StatementResults.GetHeaders()
	}
	return statement.Traits.GetColumnNames()
}

func (t *ResultFetcherOnPrem) Close() {
	t.refreshState.setState(types.Paused)
	statement := t.GetStatement()
	if statement.Status == types.RUNNING {
		go utils.WithPanicRecovery(func() {
			t.store.StopStatement(statement.StatementName)
		})()
	}
}

func (t *ResultFetcherOnPrem) SetRefreshCallback(refreshCallback func()) {
	t.refreshCallback = refreshCallback
}

func (t *ResultFetcherOnPrem) GetMaterializedStatementResults() *types.MaterializedStatementResults {
	return &t.materializedStatementResults
}

func (t *ResultFetcherOnPrem) GetLastRefreshTimestamp() *time.Time {
	return t.refreshState.getTimestamp()
}
