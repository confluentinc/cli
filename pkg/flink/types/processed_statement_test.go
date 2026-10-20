package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	testUtils "github.com/confluentinc/cli/v4/pkg/flink/test"
)

func TestPrintStatusMessagePrintsStructuredWarningsInsteadOfStatusDetail(t *testing.T) {
	statement := ProcessedStatement{
		Status:       RUNNING,
		StatusDetail: "[Warning] legacy inlined warning",
		Warnings:     []StatementWarning{{Severity: "CRITICAL", Reason: "SOME_REASON", Message: "Fix the query."}},
	}

	stdout := testUtils.RunAndCaptureSTDOUT(t, statement.PrintStatusMessage)

	require.Contains(t, stdout, "CRITICAL [SOME_REASON]")
	require.Contains(t, stdout, "Fix the query.")
	require.NotContains(t, stdout, "legacy inlined warning")
	require.NotContains(t, stdout, "Details: ")
}

func TestPrintStatusMessagePrintsStatusDetailOfFailedStatementAlongsideWarnings(t *testing.T) {
	statement := ProcessedStatement{
		Status:       FAILED,
		StatusDetail: "the failure reason",
		Warnings:     []StatementWarning{{Severity: "LOW", Reason: "SOME_REASON", Message: "Fix the query."}},
	}

	stdout := testUtils.RunAndCaptureSTDOUT(t, statement.PrintStatusMessage)

	require.Contains(t, stdout, "the failure reason")
	require.Contains(t, stdout, "LOW [SOME_REASON]")
}

func TestPrintStatusMessagePrintsStatusDetailWhenThereAreNoWarnings(t *testing.T) {
	statement := ProcessedStatement{Status: RUNNING, StatusDetail: "something worth knowing"}

	stdout := testUtils.RunAndCaptureSTDOUT(t, statement.PrintStatusMessage)

	require.Contains(t, stdout, "Details: ")
	require.Contains(t, stdout, "something worth knowing")
	require.NotContains(t, stdout, "Warnings:")
}

func TestPrintStatusMessagePrintsNoWarningsBlockWhenThereAreNoWarnings(t *testing.T) {
	statement := ProcessedStatement{Status: RUNNING}

	stdout := testUtils.RunAndCaptureSTDOUT(t, statement.PrintStatusMessage)

	require.Contains(t, stdout, "Statement successfully submitted.")
	require.NotContains(t, stdout, "Warnings:")
	require.NotContains(t, stdout, "Details: ")
}

func TestPrintOutputDryRunStatementPrintsWarnings(t *testing.T) {
	statement := ProcessedStatement{
		Status:   COMPLETED,
		Warnings: []StatementWarning{{Severity: "MODERATE", Reason: "SOME_REASON", Message: "Fix the query."}},
	}

	stdout := testUtils.RunAndCaptureSTDOUT(t, statement.PrintOutputDryRunStatement)

	require.Contains(t, stdout, "MODERATE [SOME_REASON]")
	require.Contains(t, stdout, "Fix the query.")
}
