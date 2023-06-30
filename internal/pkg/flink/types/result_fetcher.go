package types

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
	FetchNextPageAndUpdateState()
	JumpToLastPage()
	Init(statement ProcessedStatement)
	Close()
	SetAutoRefreshCallback(func())
	GetResults() *MaterializedStatementResults
}
