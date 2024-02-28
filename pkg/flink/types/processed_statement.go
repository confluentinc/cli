package types

import (
	"fmt"
	"strings"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v3/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type PHASE string

const (
	PENDING   PHASE = "PENDING"   // Results are not available yet
	RUNNING   PHASE = "RUNNING"   // More results are available (pagination)
	COMPLETED PHASE = "COMPLETED" // All results were fetched
	FAILED    PHASE = "FAILED"
)

// Custom Internal type that shall be used internally by the client
type ProcessedStatement struct {
	Statement            string `json:"statement"`
	StatementName        string `json:"statement_name"`
	Kind                 string `json:"kind"`
	ComputePool          string `json:"compute_pool"`
	Principal            string `json:"principal"`
	Status               PHASE  `json:"status"`
	StatusDetail         string `json:"status_detail,omitempty"` // Shown at the top before the table
	IsLocalStatement     bool
	IsSensitiveStatement bool
	PageToken            string
	StatementResults     *StatementResults
	Traits               flinkgatewayv1.SqlV1StatementTraits
}

func NewProcessedStatement(statementObj flinkgatewayv1.SqlV1Statement) *ProcessedStatement {
	return &ProcessedStatement{
		Statement:     statementObj.Spec.GetStatement(),
		StatementName: statementObj.GetName(),
		ComputePool:   statementObj.Spec.GetComputePoolId(),
		Principal:     statementObj.Spec.GetPrincipal(),
		StatusDetail:  statementObj.Status.GetDetail(),
		Status:        PHASE(statementObj.Status.GetPhase()),
		Traits:        statementObj.Status.GetTraits(),
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
		output.Printf(false, "Statement phase is %s.\n", s.Status)
	}
	if s.StatusDetail != "" {
		output.Printf(false, "%s.\n", s.StatusDetail)
	}
}

func (s ProcessedStatement) IsTerminalState() bool {
	isRunningAndHasNextPage := s.Status == RUNNING && s.PageToken != ""
	hasResults := len(s.StatementResults.GetRows()) > 1
	return s.Status == COMPLETED || s.Status == FAILED || isRunningAndHasNextPage || hasResults
}

func (s ProcessedStatement) IsSelectStatement() bool {
	return strings.EqualFold(s.Traits.GetSqlKind(), "SELECT")
}
