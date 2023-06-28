package types

type ApplicationControllerInterface interface {
	ExitApplication()
	AddCleanupFunction(func()) ApplicationControllerInterface
}

type InputControllerInterface interface {
	GetUserInput() string
	IsSpecialInput(string) bool
	GetWindowWidth() int
}

type TableControllerInterface interface {
	Start()
	Init(statement ProcessedStatement)
}

type StatementControllerInterface interface {
	ExecuteStatement(statementToExecute string) (*ProcessedStatement, *StatementError)
}

type OutputControllerInterface interface {
	HandleStatementResults(processedStatement ProcessedStatement, windowSize int)
}

type FetchControllerInterface interface {
	GetFetchState() FetchState
	IsTableMode() bool
	ToggleTableMode()
	ToggleAutoRefresh()
	IsAutoRefreshRunning() bool
	FetchNextPage()
	JumpToLastPage()
	Init(statement ProcessedStatement)
	Close()
	SetAutoRefreshCallback(func())
	GetResults() *MaterializedStatementResults
}
