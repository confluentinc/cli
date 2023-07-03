package controller

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/store"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type FetchController struct {
	store                        store.StoreInterface
	statement                    types.ProcessedStatement
	statementLock                sync.RWMutex
	materializedStatementResults types.MaterializedStatementResults
	fetchState                   int32
	autoRefreshCallback          func()
}

const (
	maxResultsCapacity     int  = 1000
	defaultRefreshInterval uint = 1000 // in milliseconds
	minColumnWidth         int  = 4    // min characters displayed in a column
)

func NewFetchController(store store.StoreInterface) types.FetchControllerInterface {
	return &FetchController{
		store:               store,
		autoRefreshCallback: func() {},
	}
}

func (t *FetchController) getStatement() types.ProcessedStatement {
	t.statementLock.RLock()
	defer t.statementLock.RUnlock()

	return t.statement
}

func (t *FetchController) setStatement(statement types.ProcessedStatement) {
	t.statementLock.Lock()
	defer t.statementLock.Unlock()

	t.statement = statement
}

func (t *FetchController) GetFetchState() types.FetchState {
	return types.FetchState(atomic.LoadInt32(&t.fetchState))
}

func (t *FetchController) IsTableMode() bool {
	return t.materializedStatementResults.IsTableMode()
}

func (t *FetchController) ToggleTableMode() {
	t.materializedStatementResults.SetTableMode(!t.materializedStatementResults.IsTableMode())
}

func (t *FetchController) ToggleAutoRefresh() {
	if t.IsAutoRefreshRunning() {
		t.setFetchState(types.Paused)
		return
	}

	t.startAutoRefresh(defaultRefreshInterval)
}

func (t *FetchController) IsAutoRefreshRunning() bool {
	return t.GetFetchState() == types.Running
}

func (t *FetchController) startAutoRefresh(refreshInterval uint) {
	if t.isAutoRefreshStartAllowed() {
		t.setFetchState(types.Running)
		go func() {
			for t.IsAutoRefreshRunning() {
				t.fetchNextPage()
				t.autoRefreshCallback()
				time.Sleep(time.Millisecond * time.Duration(refreshInterval))
			}
		}()
	}
}

func (t *FetchController) isAutoRefreshStartAllowed() bool {
	return t.GetFetchState() == types.Paused || t.GetFetchState() == types.Failed
}

func (t *FetchController) setFetchState(state types.FetchState) {
	atomic.StoreInt32(&t.fetchState, int32(state))
}

func (t *FetchController) fetchNextPage() {
	// don't fetch if we're already at the last page, otherwise we would fetch the first page again
	if t.GetFetchState() == types.Completed {
		return
	}

	// fetch
	newResults, err := t.store.FetchStatementResults(t.getStatement())
	if err != nil {
		t.setFetchState(types.Failed)
		return
	}

	// update data
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

func (t *FetchController) GetHeaders() []string {
	return t.materializedStatementResults.GetHeaders()
}

func (t *FetchController) GetMaxWidthPerColumn() []int {
	return t.materializedStatementResults.GetMaxWidthPerColum()
}

func (t *FetchController) GetResultsIterator(startFromBack bool) types.MaterializedStatementResultsIterator {
	return t.materializedStatementResults.Iterator(startFromBack)
}

func (t *FetchController) ForEach(f func(rowIdx int, row *types.StatementResultRow)) {
	t.materializedStatementResults.ForEach(f)
}

func (t *FetchController) Init(statement types.ProcessedStatement) {
	t.setFetchState(types.Paused)
	t.setStatement(statement)
	t.materializedStatementResults = types.NewMaterializedStatementResults(statement.StatementResults.GetHeaders(), maxResultsCapacity)
	t.materializedStatementResults.SetTableMode(true)
	t.materializedStatementResults.Append(statement.StatementResults.GetRows()...)
	// if unbounded result start refreshing results in the background
	if statement.PageToken != "" {
		t.startAutoRefresh(defaultRefreshInterval)
	} else {
		t.setFetchState(types.Completed)
	}
}

func (t *FetchController) Close() {
	t.setFetchState(types.Paused)
	// This was used to delete statements after their execution to save system resources, which should not be
	// an issue anymore. We don't want to remove it completely just yet, but will disable it by default for now.
	// TODO: remove this completely once we are sure we won't need it in the future
	statement := t.getStatement()
	if config.ShouldCleanupStatements || statement.Status == types.RUNNING {
		go t.store.DeleteStatement(statement.StatementName)
	}
}

func (t *FetchController) SetAutoRefreshCallback(autoRefreshCallback func()) {
	t.autoRefreshCallback = autoRefreshCallback
}
