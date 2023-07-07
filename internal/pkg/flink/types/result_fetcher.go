package types

import "time"

type FetchState int32

const (
	Paused    FetchState = iota // auto fetch was Paused
	Completed                   // arrived at last page, fetch is Completed
	Failed                      // fetching next page Failed
	Running                     // auto fetch is Running
)

type ResultFetcherInterface interface {
	GetFetchState() FetchState
	IsTableMode() bool
	ToggleTableMode()
	ToggleAutoRefresh()
	IsAutoRefreshRunning() bool
	Init(statement ProcessedStatement)
	Close()
	SetAutoRefreshCallback(func())
	GetStatement() ProcessedStatement
	GetMaterializedStatementResults() *MaterializedStatementResults
	GetLastFetchTimestamp() *time.Time
}
