package types

import (
	"fmt"
	"strings"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/output"
	cliutils "github.com/confluentinc/cli/internal/pkg/utils"
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

func (r *StatementResultRow) GetRowKey() string {
	rowKey := strings.Builder{}
	for _, field := range r.GetFields() {
		rowKey.WriteString(field.ToString())
	}
	return rowKey.String()
}

func (r *StatementResultRow) GetFields() []StatementResultField {
	if r == nil {
		var fields []StatementResultField
		return fields
	}
	return r.Fields
}

const (
	INSERT        StatementResultOperation = 0
	UPDATE_BEFORE StatementResultOperation = 1
	UPDATE_AFTER  StatementResultOperation = 2
	DELETE        StatementResultOperation = 3
)

type StatementResultOperation float64

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
	Message          string
	HttpResponseCode int
	FailureMessage   string
	Usage            []string
	Suggestion       string
}

func (e *StatementError) Error() string {
	if e == nil {
		return ""
	}
	errStr := "Error: no message"
	if e.Message != "" {
		errStr = fmt.Sprintf("Error: %s", e.Message)
	}
	if len(e.Usage) > 0 {
		errStr += fmt.Sprintf("\nUsage: %s", cliutils.ArrayToCommaDelimitedString(e.Usage, "or"))
	}
	if e.Suggestion != "" {
		errStr += fmt.Sprintf("\nSuggestion: %s", e.Suggestion)
	}
	if e.FailureMessage != "" {
		errStr += fmt.Sprintf("\nError details: %s", e.FailureMessage)
	}

	return errStr
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
	StatementName     string `json:"statement_name"`
	Kind              string `json:"statement"`
	ComputePool       string `json:"compute_pool"`
	Status            PHASE  `json:"status"`
	StatusDetail      string `json:"status_detail,omitempty"` // Shown at the top before the table
	IsLocalStatement  bool
	IsSelectStatement bool
	PageToken         string
	ResultSchema      flinkgatewayv1alpha1.SqlV1alpha1ResultSchema
	StatementResults  *StatementResults
}

func NewProcessedStatement(statementObj flinkgatewayv1alpha1.SqlV1alpha1Statement) *ProcessedStatement {
	statement := strings.ToLower(strings.TrimSpace(statementObj.Spec.GetStatement()))
	return &ProcessedStatement{
		StatementName:     statementObj.Spec.GetStatementName(),
		StatusDetail:      statementObj.Status.GetDetail(),
		Status:            PHASE(statementObj.Status.GetPhase()),
		ResultSchema:      statementObj.Status.GetResultSchema(),
		IsSelectStatement: strings.HasPrefix(statement, "select"),
	}
}

func (s ProcessedStatement) PrintStatusMessage() {
	if s.IsLocalStatement {
		s.printStatusMessageOfLocalStatement()
	} else {
		s.printStatusMessageOfNonLocalStatement()
	}
}

func (s ProcessedStatement) printStatusMessageOfLocalStatement() {
	if s.Status == "FAILED" {
		utils.OutputErr(fmt.Sprintf("Error: %s", "couldn't process statement, please check your statement and try again"))
	} else {
		utils.OutputInfo("Statement successfully submitted.")
	}
}

func (s ProcessedStatement) printStatusMessageOfNonLocalStatement() {

	if s.StatementName != "" {
		utils.OutputInfof("Statement name: %s\n", s.StatementName)
	}
	if s.Status == "FAILED" {
		utils.OutputErr(fmt.Sprintf("Error: %s", "statement submission failed"))
	} else {
		utils.OutputInfo("Statement successfully submitted.")
		utils.OutputInfo(fmt.Sprintf("Waiting for statement to be ready. Statement phase is %s.", s.Status))
	}
}

func (s ProcessedStatement) GetPageSize() int {
	return len(s.StatementResults.GetRows())
}

func (s ProcessedStatement) PrintStatementDoneStatus() {
	if s.Status != "" {
		output.Printf("Statement phase is %s.\n", s.Status)
	}
	if s.StatusDetail != "" {
		output.Printf("%s.\n", s.StatusDetail)
	}
}
