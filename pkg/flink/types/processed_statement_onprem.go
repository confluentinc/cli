package types

import (
	"fmt"
	"strings"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v4/pkg/output"
)

// ProcessedStatementOnPrem Custom Internal type that shall be used internally by the client
type ProcessedStatementOnPrem struct {
	Statement            string `json:"statement"`
	StatementName        string `json:"statement_name"`
	Kind                 string `json:"kind"`
	ComputePool          string `json:"compute_pool"`
	Status               PHASE  `json:"status"`
	StatusDetail         string `json:"status_detail,omitempty"` // Shown at the top before the table
	IsLocalStatement     bool
	IsSensitiveStatement bool
	PageToken            string
	Properties           map[string]string
	StatementResults     *StatementResults // TODO: maybe this needs a specific CMF type
	Traits               cmfsdk.StatementTraits
}

func NewProcessedStatementOnPrem(statementObj cmfsdk.Statement) *ProcessedStatementOnPrem {
	return &ProcessedStatementOnPrem{
		Statement:     statementObj.Spec.GetStatement(),
		StatementName: statementObj.Metadata.GetName(),
		ComputePool:   statementObj.Spec.GetComputePoolName(),
		StatusDetail:  statementObj.Status.GetDetail(),
		Status:        PHASE(statementObj.Status.GetPhase()),
		Properties:    statementObj.Spec.GetProperties(),
		Traits:        statementObj.Status.GetTraits(),
	}
}

func (s ProcessedStatementOnPrem) PrintStatusMessage() {
	if s.IsLocalStatement {
		s.printStatusMessageOfLocalStatement()
	} else {
		s.printStatusMessageOfNonLocalStatement()
	}
}

func (s ProcessedStatementOnPrem) printStatusMessageOfLocalStatement() {
	if s.Status == "FAILED" {
		utils.OutputErr(fmt.Sprintf("Error: %s", "couldn't process statement, please check your statement and try again"))
	} else {
		utils.OutputInfo("Statement successfully submitted.")
	}
}

func (s ProcessedStatementOnPrem) printStatusMessageOfNonLocalStatement() {
	if s.Status == "FAILED" {
		utils.OutputErr(fmt.Sprintf("Error: %s", "statement submission failed"))
	} else {
		utils.OutputInfo("Statement successfully submitted.")
		utils.OutputInfo(fmt.Sprintf("Waiting for statement to be ready. Statement phase is %s.", s.Status))
	}
}

func (s ProcessedStatementOnPrem) PrintOutputDryRunStatement() {
	utils.OutputInfo(fmt.Sprintf("Statement successfully submitted. Statement phase is %s.", s.Status))
	if s.Status == "FAILED" {
		utils.OutputErr(fmt.Sprintf("Dry run statement was verified and there were issues found.\nError: %s", s.StatusDetail))
	} else if s.Status == "COMPLETED" {
		utils.OutputInfo("Dry run statement was verified and there were no issues found.")
		utils.OutputWarn("If you wish to submit your statement, disable dry run mode before submitting your statement with \"set 'sql.dry-run' = 'false';\"")
	} else {
		utils.OutputErr(fmt.Sprintf("Dry run statement execution resulted in unexpected status.\nStatus: %s", s.Status))
		utils.OutputErr(fmt.Sprintf("Details: %s", s.StatusDetail))
	}
}

func (s ProcessedStatementOnPrem) GetPageSize() int {
	return len(s.StatementResults.GetRows())
}

func (s ProcessedStatementOnPrem) PrintStatementDoneStatus() {
	if s.Status != "" {
		output.Printf(false, "Statement phase is %s.\n", s.Status)
	}
	if s.StatusDetail != "" {
		output.Printf(false, "%s.\n", s.StatusDetail)
	}
}

func (s ProcessedStatementOnPrem) IsTerminalState() bool {
	isRunningAndHasNextPage := s.Status == RUNNING && s.PageToken != ""
	hasResults := len(s.StatementResults.GetRows()) > 1
	return s.Status == COMPLETED || s.Status == FAILED || isRunningAndHasNextPage || hasResults
}

func (s ProcessedStatementOnPrem) IsSelectStatement() bool {
	return strings.EqualFold(s.Traits.GetSqlKind(), "SELECT") ||
		strings.HasPrefix(strings.ToUpper(s.Statement), "SELECT")
}

func (s ProcessedStatementOnPrem) IsDryRunStatement() bool {
	keyVal, ok := s.Properties[config.KeyDryRun]

	if ok && strings.ToLower(keyVal) == "true" {
		return true
	}
	return false
}
