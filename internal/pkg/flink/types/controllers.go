package types

type ApplicationControllerInterface interface {
	ExitApplication()
	AddCleanupFunction(func()) ApplicationControllerInterface
}

type InputControllerInterface interface {
	GetUserInput() string
	HasUserInitiatedExit(userInput string) bool
	HasUserEnabledReverseSearch() bool
	StartReverseSearch()
	GetWindowWidth() int
}

type StatementControllerInterface interface {
	ExecuteStatement(statementToExecute string) (*ProcessedStatement, *StatementError)
}

type OutputControllerInterface interface {
	VisualizeResults()
}
