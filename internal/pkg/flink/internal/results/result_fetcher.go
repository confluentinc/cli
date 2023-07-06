package results

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type ResultFetcher struct {
	store                        types.StoreInterface
	statement                    types.ProcessedStatement
	statementLock                sync.RWMutex
	materializedStatementResults types.MaterializedStatementResults
	fetchState                   int32
	autoRefreshCallback          func()
}

const (
	MaxResultsCapacity     int  = 10000
	DefaultRefreshInterval uint = 1000 // in milliseconds
)

func NewResultFetcher(store types.StoreInterface) types.ResultFetcherInterface {
	return &ResultFetcher{
		store:               store,
		autoRefreshCallback: func() {},
	}
}

func (t *ResultFetcher) IsTableMode() bool {
	return t.materializedStatementResults.IsTableMode()
}

func (t *ResultFetcher) ToggleTableMode() {
	t.materializedStatementResults.SetTableMode(!t.materializedStatementResults.IsTableMode())
}

func (t *ResultFetcher) ToggleAutoRefresh() {
	if t.IsAutoRefreshRunning() {
		t.setFetchState(types.Paused)
		return
	}

	t.startAutoRefresh(DefaultRefreshInterval)
}

func (t *ResultFetcher) IsAutoRefreshRunning() bool {
	return t.GetFetchState() == types.Running
}

func (t *ResultFetcher) GetFetchState() types.FetchState {
	return types.FetchState(atomic.LoadInt32(&t.fetchState))
}

func (t *ResultFetcher) setFetchState(state types.FetchState) {
	atomic.StoreInt32(&t.fetchState, int32(state))
}

func (t *ResultFetcher) startAutoRefresh(refreshInterval uint) {
	if t.isAutoRefreshStartAllowed() {
		t.setFetchState(types.Running)
		go func() {
			for t.IsAutoRefreshRunning() {
				t.fetchNextPageAndUpdateState()
				t.autoRefreshCallback()
				time.Sleep(time.Millisecond * time.Duration(refreshInterval))
			}
		}()
	}
}

func (t *ResultFetcher) isAutoRefreshStartAllowed() bool {
	return t.GetFetchState() == types.Paused || t.GetFetchState() == types.Failed
}

func (t *ResultFetcher) fetchNextPageAndUpdateState() {
	newResults, err := t.store.FetchStatementResults(t.GetStatement())
	t.updateState(newResults, err)
}

func (t *ResultFetcher) updateState(newResults *types.ProcessedStatement, err *types.StatementError) {
	// don't fetch if we're already at the last page, otherwise we would fetch the first page again
	if t.GetFetchState() == types.Completed {
		return
	}

	if err != nil {
		t.setFetchState(types.Failed)
		return
	}

	t.setStatement(*newResults)
	t.materializedStatementResults.Append(newResults.StatementResults.GetRows()...)
	if newResults.PageToken == "" {
		t.setFetchState(types.Completed)
		return
	}

	// if auto refresh is not running we set the state to types.Paused
	if !t.IsAutoRefreshRunning() {
		t.setFetchState(types.Paused)
		return
	}

	t.setFetchState(types.Running)
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
	t.setInitialFetchState(statement)
	headers := t.getResultHeadersOrCreateFromResultSchema(statement)
	t.materializedStatementResults = types.NewMaterializedStatementResults(headers, MaxResultsCapacity)
	t.materializedStatementResults.SetTableMode(true)
	t.materializedStatementResults.Append(statement.StatementResults.GetRows()...)
}

func (t *ResultFetcher) setInitialFetchState(statement types.ProcessedStatement) {
	if statement.PageToken == "" {
		t.setFetchState(types.Completed)
		return
	}
	t.setFetchState(types.Paused)
}

func (t *ResultFetcher) getResultHeadersOrCreateFromResultSchema(statement types.ProcessedStatement) []string {
	if len(statement.StatementResults.GetHeaders()) > 0 {
		return statement.StatementResults.GetHeaders()
	}
	headers := make([]string, len(statement.ResultSchema.GetColumns()))
	for idx, column := range statement.ResultSchema.GetColumns() {
		headers[idx] = column.GetName()
	}
	return headers
}

func (t *ResultFetcher) Close() {
	t.setFetchState(types.Paused)
	// This was used to delete statements after their execution to save system resources, which should not be
	// an issue anymore. We don't want to remove it completely just yet, but will disable it by default for now.
	// TODO: remove this completely once we are sure we won't need it in the future
	statement := t.GetStatement()
	if config.ShouldCleanupStatements || statement.Status == types.RUNNING {
		go t.store.DeleteStatement(statement.StatementName)
	}
}

func (t *ResultFetcher) SetAutoRefreshCallback(autoRefreshCallback func()) {
	t.autoRefreshCallback = autoRefreshCallback
}

func (t *ResultFetcher) GetMaterializedStatementResults() *types.MaterializedStatementResults {
	return &t.materializedStatementResults
}
