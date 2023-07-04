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

type StatementControllerInterface interface {
	ExecuteStatement(statementToExecute string) (*ProcessedStatement, *StatementError)
}

type OutputControllerInterface interface {
	VisualizeResults()
}
