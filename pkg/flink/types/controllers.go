package types

import "github.com/sourcegraph/go-lsp"

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
	SetDiagnostics(diagnostics []lsp.Diagnostic)
	DiagnosticsEnabled() bool
}

type StatementControllerInterface interface {
	ExecuteStatement(statementToExecute string) (*ProcessedStatement, *StatementError)
	CleanupStatement()
}

type OutputControllerInterface interface {
	VisualizeResults()
}
