package types

import "time"

type RefreshState int32

const (
	Paused    RefreshState = iota // auto fetch was Paused
	Completed                     // arrived at last page, fetch is Completed
	Failed                        // fetching next page Failed
	Running                       // auto fetch is Running
)

func (f RefreshState) ToString() string {
	switch f {
	case Completed:
		return "Completed"
	case Failed:
		return "Failed"
	case Paused:
		return "Paused"
	case Running:
		return "Running"
	}
	return "Unknown"
}

type ResultFetcherInterface interface {
	GetFetchState() RefreshState
	IsTableMode() bool
	ToggleTableMode()
	ToggleRefresh()
	IsRefreshRunning() bool
	Init(statement ProcessedStatement)
	Close()
	SetAutoRefreshCallback(func())
	GetStatement() ProcessedStatement
	GetMaterializedStatementResults() *MaterializedStatementResults
	GetLastFetchTimestamp() *time.Time
}
