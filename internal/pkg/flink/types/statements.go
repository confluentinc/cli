package types

import (
	"fmt"
	"strings"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"
)

type StatementResults struct {
	Headers []string
	Rows    []StatementResultRow
}

func (s *StatementResults) GetHeaders() []string {
	if s == nil {
		return []string{}
	}
	return s.Headers
}

func (s *StatementResults) GetRows() []StatementResultRow {
	if s == nil {
		return []StatementResultRow{}
	}
	return s.Rows
}

type StatementResultRow struct {
	Operation StatementResultOperation
	Fields    []StatementResultField
}

func (r StatementResultRow) GetRowKey() string {
	rowKey := strings.Builder{}
	for _, field := range r.Fields {
		rowKey.WriteString(field.ToString())
	}
	return rowKey.String()
}

const (
	INSERT        StatementResultOperation = 0
	UPDATE_BEFORE StatementResultOperation = 1
	UPDATE_AFTER  StatementResultOperation = 2
	DELETE        StatementResultOperation = 3
)

type StatementResultOperation int8

func (s StatementResultOperation) IsInsertOperation() bool {
	return s == INSERT || s == UPDATE_AFTER
}

func (s StatementResultOperation) String() string {
	switch s {
	case INSERT:
		return "+I"
	case UPDATE_BEFORE:
		return "-U"
	case UPDATE_AFTER:
		return "+U"
	case DELETE:
		return "-D"
	}
	return ""
}

type MockStatementResult struct {
	ResultSchema     flinkgatewayv1alpha1.SqlV1alpha1ResultSchema
	StatementResults flinkgatewayv1alpha1.SqlV1alpha1StatementResult
}

type StatementError struct {
	Msg              string
	HttpResponseCode int
	FailureMessage   string
}

func (e *StatementError) Error() string {
	if e == nil {
		return ""
	}
	if e.Msg == "" {
		return "Error"
	}
	if e.FailureMessage == "" {
		return fmt.Sprintf("Error: %s", e.Msg)
	}
	return fmt.Sprintf("Error: %s \nError details: %s", e.Msg, e.FailureMessage)
}

type PHASE string

const (
	PENDING   PHASE = "PENDING"   // Results are not available yet
	RUNNING   PHASE = "RUNNING"   // More results are available (pagination)
	COMPLETED PHASE = "COMPLETED" // All results were fetched
	DELETING  PHASE = "DELETING"
	FAILED    PHASE = "FAILED"
	FAILING   PHASE = "FAILING"
)

// Custom Internal type that shall be used internally by the client
type ProcessedStatement struct {
	StatementName    string `json:"statement_name"`
	Kind             string `json:"statement"`
	ComputePool      string `json:"compute_pool"`
	Status           PHASE  `json:"status"`
	StatusDetail     string `json:"status_detail,omitempty"` // Shown at the top before the table
	IsLocalStatement bool
	PageToken        string
	ResultSchema     flinkgatewayv1alpha1.SqlV1alpha1ResultSchema
	StatementResults *StatementResults
}

func NewProcessedStatement(statementObj flinkgatewayv1alpha1.SqlV1alpha1Statement) *ProcessedStatement {
	return &ProcessedStatement{
		StatementName: statementObj.Spec.GetStatementName(),
		StatusDetail:  statementObj.Status.GetDetail(),
		Status:        PHASE(statementObj.Status.GetPhase()),
		ResultSchema:  statementObj.Status.GetResultSchema(),
	}
}
