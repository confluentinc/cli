package types

import (
	"fmt"
	"strings"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/output"
)

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

func (s ProcessedStatement) IsTerminalState() bool {
	isRunningAndHasResults := s.Status == RUNNING && s.PageToken != ""
	return s.Status == COMPLETED || s.Status == FAILED || isRunningAndHasResults
}
