package types

import "context"

type StoreInterface interface {
	ProcessStatement(statement string) (*ProcessedStatement, *StatementError)
	FetchStatementResults(ProcessedStatement) (*ProcessedStatement, *StatementError)
	StopStatement(statementName string) bool
	DeleteStatement(statementName string) bool
	WaitPendingStatement(ctx context.Context, statement ProcessedStatement) (*ProcessedStatement, *StatementError)
	WaitForTerminalStatementState(ctx context.Context, statement ProcessedStatement) (*ProcessedStatement, *StatementError)
	GetCurrentCatalog() string
	GetCurrentDatabase() string
}
