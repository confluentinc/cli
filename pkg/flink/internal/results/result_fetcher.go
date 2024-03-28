package results

import (
	"sync"
	"time"

	"github.com/confluentinc/cli/v3/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
)

type refreshState struct {
	mutex     sync.RWMutex
	timestamp *time.Time
	state     types.RefreshState
}

func (s *refreshState) getTimestamp() *time.Time {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.timestamp
}

func (s *refreshState) getState() types.RefreshState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.state
}

func (s *refreshState) setState(state types.RefreshState) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state = state
	now := time.Now()
	s.timestamp = &now
}

type ResultFetcher struct {
	store                        types.StoreInterface
	statement                    types.ProcessedStatement
	statementLock                sync.RWMutex
	materializedStatementResults types.MaterializedStatementResults
	refreshState                 refreshState
	refreshCallback              func()
	fetchLock                    sync.Mutex
}

const (
	MaxResultsCapacity     int  = 10000
	DefaultRefreshInterval uint = 1000 // in milliseconds
)

func NewResultFetcher(store types.StoreInterface) types.ResultFetcherInterface {
	return &ResultFetcher{
		store:           store,
		refreshCallback: func() {},
	}
}

func (t *ResultFetcher) IsTableMode() bool {
	return t.materializedStatementResults.IsTableMode()
}

func (t *ResultFetcher) ToggleTableMode() {
	t.materializedStatementResults.SetTableMode(!t.materializedStatementResults.IsTableMode())
}

func (t *ResultFetcher) ToggleRefresh() {
	if t.IsRefreshRunning() {
		t.refreshState.setState(types.Paused)
		return
	}

	t.startRefresh(DefaultRefreshInterval)
}

func (t *ResultFetcher) IsRefreshRunning() bool {
	return t.GetRefreshState() == types.Running
}

func (t *ResultFetcher) GetRefreshState() types.RefreshState {
	return t.refreshState.getState()
}

func (t *ResultFetcher) startRefresh(refreshInterval uint) {
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

func (t *ResultFetcher) isRefreshStartAllowed() bool {
	return t.GetRefreshState() == types.Paused || t.GetRefreshState() == types.Failed
}

func (t *ResultFetcher) fetchNextPageAndUpdateState() {
	// lock here to make sure we don't fetch the same page twice
	t.fetchLock.Lock()
	defer t.fetchLock.Unlock()

	newResults, err := t.store.FetchStatementResults(t.GetStatement())
	t.updateState(newResults, err)
}

func (t *ResultFetcher) updateState(newResults *types.ProcessedStatement, err *types.StatementError) {
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

func (t *ResultFetcher) GetStatement() types.ProcessedStatement {
	t.statementLock.RLock()
	defer t.statementLock.RUnlock()

	return t.statement
}

func (t *ResultFetcher) setStatement(statement types.ProcessedStatement) {
	t.statementLock.Lock()
	defer t.statementLock.Unlock()

	t.statement = statement
}

func (t *ResultFetcher) Init(statement types.ProcessedStatement) {
	t.setStatement(statement)
	t.setInitialRefreshState(statement)
	headers := t.getResultHeadersOrCreateFromResultSchema(statement)
	t.materializedStatementResults = types.NewMaterializedStatementResults(headers, MaxResultsCapacity, statement.Traits.UpsertColumns)
	t.materializedStatementResults.SetTableMode(true)
	t.materializedStatementResults.Append(statement.StatementResults.GetRows()...)
}

func (t *ResultFetcher) setInitialRefreshState(statement types.ProcessedStatement) {
	if statement.PageToken == "" {
		t.refreshState.setState(types.Completed)
		return
	}
	t.refreshState.setState(types.Paused)
}

func (t *ResultFetcher) getResultHeadersOrCreateFromResultSchema(statement types.ProcessedStatement) []string {
	if len(statement.StatementResults.GetHeaders()) > 0 {
		return statement.StatementResults.GetHeaders()
	}
	headers := make([]string, len(statement.Traits.Schema.GetColumns()))
	for idx, column := range statement.Traits.Schema.GetColumns() {
		headers[idx] = column.GetName()
	}
	return headers
}

func (t *ResultFetcher) Close() {
	t.refreshState.setState(types.Paused)
	statement := t.GetStatement()
	if statement.Status == types.RUNNING {
		go utils.WithPanicRecovery(func() {
			t.store.StopStatement(statement.StatementName)
		})()
	}
}

func (t *ResultFetcher) SetRefreshCallback(refreshCallback func()) {
	t.refreshCallback = refreshCallback
}

func (t *ResultFetcher) GetMaterializedStatementResults() *types.MaterializedStatementResults {
	return &t.materializedStatementResults
}

func (t *ResultFetcher) GetLastRefreshTimestamp() *time.Time {
	return t.refreshState.getTimestamp()
}
