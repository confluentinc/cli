package types

import (
	"fmt"
	"strings"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type PHASE string

const (
	PENDING   PHASE = "PENDING"   // Results are not available yet
	RUNNING   PHASE = "RUNNING"   // More results are available (pagination)
	COMPLETED PHASE = "COMPLETED" // All results were fetched
	FAILED    PHASE = "FAILED"
)

// ProcessedStatement Custom Internal type that shall be used internally by the client
type ProcessedStatement struct {
	Statement            string `json:"statement"`
	StatementName        string `json:"statement_name"`
	Kind                 string `json:"kind"`
	ComputePool          string `json:"compute_pool"`
	Principal            string `json:"principal"` // Cloud only
	Status               PHASE  `json:"status"`
	StatusDetail         string `json:"status_detail,omitempty"` // Shown at the top before the table
	IsLocalStatement     bool
	IsSensitiveStatement bool
	PageToken            string
	Properties           map[string]string
	StatementResults     *StatementResults
	Traits               StatementTraits
}

func NewProcessedStatement(statementObj flinkgatewayv1.SqlV1Statement) *ProcessedStatement {
	traits := statementObj.Status.GetTraits()
	return &ProcessedStatement{
		Statement:     statementObj.Spec.GetStatement(),
		StatementName: statementObj.GetName(),
		ComputePool:   statementObj.Spec.GetComputePoolId(),
		Principal:     statementObj.Spec.GetPrincipal(),
		StatusDetail:  statementObj.Status.GetDetail(),
		Status:        PHASE(statementObj.Status.GetPhase()),
		Properties:    statementObj.Spec.GetProperties(),
		Traits:        StatementTraits{FlinkGatewayV1StatementTraits: &traits},
	}
}

func NewProcessedStatementOnPrem(statementObj cmfsdk.Statement) *ProcessedStatement {
	traits := statementObj.Status.GetTraits()
	return &ProcessedStatement{
		Statement:     statementObj.Spec.GetStatement(),
		StatementName: statementObj.Metadata.GetName(),
		ComputePool:   statementObj.Spec.GetComputePoolName(),
		StatusDetail:  statementObj.Status.GetDetail(),
		Status:        PHASE(statementObj.Status.GetPhase()),
		Properties:    statementObj.Spec.GetProperties(),
		Traits:        StatementTraits{CmfStatementTraits: &traits},
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
	if s.Status == "FAILED" {
		utils.OutputErr(fmt.Sprintf("Error: %s", "statement submission failed"))
	} else {
		utils.OutputInfo("Statement successfully submitted.")

		if s.Status != "COMPLETED" {
			utils.OutputInfo(fmt.Sprintf("Waiting for statement to be ready. Statement phase: %s.", s.Status))
		}
	}

	if s.StatusDetail != "" {
		utils.OutputInfof("Details: ")
		utils.OutputWarn(s.StatusDetail)
	}
}

func (s ProcessedStatement) PrintOutputDryRunStatement() {
	utils.OutputInfo(fmt.Sprintf("Statement successfully submitted. Statement phase: %s.", s.Status))
	if s.Status == "FAILED" {
		utils.OutputErr(fmt.Sprintf("Dry run statement was verified and there were issues found.\nError: %s", s.StatusDetail))
	} else if s.Status == "COMPLETED" {
		utils.OutputInfo("Dry run statement was verified and there were no issues found.")
		utils.OutputWarn("If you wish to submit your statement, disable dry run mode before submitting your statement with \"set 'sql.dry-run' = 'false';\"")
	} else {
		utils.OutputErr(fmt.Sprintf("Dry run statement execution resulted in unexpected status.\nStatus: %s", s.Status))
		utils.OutputInfof("Details: ")
		utils.OutputErr(s.StatusDetail)
	}
}

func (s ProcessedStatement) GetPageSize() int {
	return len(s.StatementResults.GetRows())
}

func (s ProcessedStatement) PrintStatementDoneStatus() {
	if s.Status != "" {
		output.Printf(false, "Finished statement execution. Statement phase: %s.\n", s.Status)
	}
	if s.StatusDetail != "" {
		output.Printf(false, "Details: %s.\n", strings.TrimSuffix(s.StatusDetail, "."))
	}
}

func (s ProcessedStatement) IsTerminalState() bool {
	isRunningAndHasNextPage := s.Status == RUNNING && s.PageToken != ""
	hasResults := len(s.StatementResults.GetRows()) > 1
	return s.Status == COMPLETED || s.Status == FAILED || isRunningAndHasNextPage || hasResults
}

func (s ProcessedStatement) IsSelectStatement() bool {
	return strings.EqualFold(s.Traits.GetSqlKind(), "SELECT") ||
		strings.HasPrefix(strings.ToUpper(s.Statement), "SELECT")
}

func (s ProcessedStatement) IsDryRunStatement() bool {
	keyVal, ok := s.Properties[config.KeyDryRun]

	if ok && strings.ToLower(keyVal) == "true" {
		return true
	}
	return false
}
