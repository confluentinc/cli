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
	MaxResultsCapacity     int  = 1000
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
				_, _ = t.FetchNextPageAndUpdateState()
				t.autoRefreshCallback()
				time.Sleep(time.Millisecond * time.Duration(refreshInterval))
			}
		}()
	}
}

func (t *ResultFetcher) isAutoRefreshStartAllowed() bool {
	return t.GetFetchState() == types.Paused || t.GetFetchState() == types.Failed
}

func (t *ResultFetcher) FetchNextPageAndUpdateState() (*types.ProcessedStatement, *types.StatementError) {
	newResults, err := t.store.FetchStatementResults(t.getStatement())
	t.updateState(newResults, err)
	return newResults, err
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

func (t *ResultFetcher) getStatement() types.ProcessedStatement {
	t.statementLock.RLock()
	defer t.statementLock.RUnlock()

	return t.statement
}

func (t *ResultFetcher) setStatement(statement types.ProcessedStatement) {
	t.statementLock.Lock()
	defer t.statementLock.Unlock()

	t.statement = statement
}

func (t *ResultFetcher) JumpToLastPage() {
	for {
		_, _ = t.FetchNextPageAndUpdateState()
		if !t.hasMoreResults() {
			break
		}
		// minimal wait to avoid rate limiting
		time.Sleep(time.Millisecond * 50)
	}
}

func (t *ResultFetcher) hasMoreResults() bool {
	return len(t.getStatement().StatementResults.GetRows()) > 0 && t.GetFetchState() != types.Failed && t.GetFetchState() != types.Completed
}

func (t *ResultFetcher) Init(statement types.ProcessedStatement) {
	t.setFetchState(types.Paused)
	t.setStatement(statement)
	headers := t.getResultHeadersOrCreateFromResultSchema()
	t.materializedStatementResults = types.NewMaterializedStatementResults(headers, MaxResultsCapacity)
	t.materializedStatementResults.SetTableMode(true)
	t.materializedStatementResults.Append(statement.StatementResults.GetRows()...)
}

func (t *ResultFetcher) getResultHeadersOrCreateFromResultSchema() []string {
	statement := t.getStatement()
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
	statement := t.getStatement()
	if config.ShouldCleanupStatements || statement.Status == types.RUNNING {
		go t.store.DeleteStatement(statement.StatementName)
	}
}

func (t *ResultFetcher) SetAutoRefreshCallback(autoRefreshCallback func()) {
	t.autoRefreshCallback = autoRefreshCallback
}

func (t *ResultFetcher) GetResults() *types.MaterializedStatementResults {
	return &t.materializedStatementResults
}
