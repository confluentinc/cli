package types

import (
	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"
)

type StatementResults struct {
	Headers []string
	Rows    []StatementResultRow
}

type StatementResultRow struct {
	Fields []StatementResultField
}

type MockStatementResult struct {
	ResultSchema     v1.SqlV1alpha1ResultSchema
	StatementResults v1.SqlV1alpha1StatementResult
}

type StatementError struct {
	Msg              string
	HttpResponseCode int
}

func (e *StatementError) Error() string {
	return e.Msg
}

type PHASE string

const (
	PENDING   PHASE = "PENDING"   // Results are not available yet
	RUNNING   PHASE = "RUNNING"   // More results are available (pagination)
	COMPLETED PHASE = "COMPLETED" // All results were fetched
	DELETING  PHASE = "DELETING"
	FAILED    PHASE = "FAILED"
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
	ResultSchema     v1.SqlV1alpha1ResultSchema
	StatementResults *StatementResults
}

func NewProcessedStatement(statementObj v1.SqlV1alpha1Statement) *ProcessedStatement {
	return &ProcessedStatement{
		StatementName: statementObj.Spec.GetStatementName(),
		StatusDetail:  statementObj.Status.GetDetail(),
		Status:        PHASE(statementObj.Status.GetPhase()),
		ResultSchema:  statementObj.Status.GetResultSchema(),
	}
}
