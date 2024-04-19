package types

import (
	"context"

	"github.com/confluentinc/cli/v3/pkg/flink/config"
)

type StoreInterface interface {
	ProcessStatement(statement string) (*ProcessedStatement, *StatementError)
	FetchStatementResults(ProcessedStatement) (*ProcessedStatement, *StatementError)
	StopStatement(statementName string) bool
	DeleteStatement(statementName string) bool
	WaitPendingStatement(ctx context.Context, statement ProcessedStatement) (*ProcessedStatement, *StatementError)
	WaitForTerminalStatementState(ctx context.Context, statement ProcessedStatement) (*ProcessedStatement, *StatementError)
	GetCurrentCatalog() string
	GetCurrentDatabase() string
	GetOutputFormat() config.OutputFormat
}
