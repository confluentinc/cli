package types

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
)

func ptrTime(t time.Time) *time.Time {
	return &t
}

func TestNewStatementWarningsReturnsNilWhenThereAreNoWarnings(t *testing.T) {
	require.Nil(t, NewStatementWarnings(nil))
	require.Nil(t, NewStatementWarnings([]flinkgatewayv1.SqlV1StatementWarning{}))
}

func TestNewStatementWarningsSortsMostSevereFirst(t *testing.T) {
	warnings := NewStatementWarnings([]flinkgatewayv1.SqlV1StatementWarning{
		{Severity: "LOW", Reason: "LOW_ONE"},
		{Severity: "CRITICAL", Reason: "CRITICAL_ONE"},
		{Severity: "MODERATE", Reason: "MODERATE_ONE"},
		{Severity: "CRITICAL", Reason: "CRITICAL_TWO"},
	})

	reasons := make([]string, len(warnings))
	for i, warning := range warnings {
		reasons[i] = warning.Reason
	}

	require.Equal(t, []string{"CRITICAL_ONE", "CRITICAL_TWO", "MODERATE_ONE", "LOW_ONE"}, reasons)
}

func TestNewStatementWarningsSortsUnrecognizedSeverityLast(t *testing.T) {
	warnings := NewStatementWarnings([]flinkgatewayv1.SqlV1StatementWarning{
		{Severity: "SOMETHING_NEW", Reason: "NEW_ONE"},
		{Severity: "LOW", Reason: "LOW_ONE"},
	})

	require.Equal(t, "LOW_ONE", warnings[0].Reason)
	require.Equal(t, "NEW_ONE", warnings[1].Reason)
}

func TestNewStatementWarningsCopiesEveryField(t *testing.T) {
	createdAt := time.Date(2026, 7, 30, 9, 15, 0, 0, time.UTC)

	warnings := NewStatementWarnings([]flinkgatewayv1.SqlV1StatementWarning{{
		Severity:  "CRITICAL",
		Reason:    "HIGH_STATE_OPERATOR_WITHOUT_TTL",
		Message:   "Your query includes one or more highly state-intensive operators.",
		CreatedAt: createdAt,
	}})

	require.Equal(t, []StatementWarning{{
		Severity:  "CRITICAL",
		Reason:    "HIGH_STATE_OPERATOR_WITHOUT_TTL",
		Message:   "Your query includes one or more highly state-intensive operators.",
		CreatedAt: &createdAt,
	}}, warnings)
}

func TestNewStatementWarningsLeavesAnAbsentTimestampUnset(t *testing.T) {
	warnings := NewStatementWarnings([]flinkgatewayv1.SqlV1StatementWarning{
		{Severity: "LOW", Reason: "SOME_REASON", Message: "Something to know."},
	})

	require.Nil(t, warnings[0].CreatedAt)

	serialized, err := json.Marshal(warnings)
	require.NoError(t, err)
	require.NotContains(t, string(serialized), "created_at")
	require.NotContains(t, string(serialized), "0001-01-01")
}

func TestFormatStatementWarningsReturnsEmptyStringWhenThereAreNoWarnings(t *testing.T) {
	require.Empty(t, FormatStatementWarnings(nil))
	require.Empty(t, FormatStatementWarnings([]StatementWarning{}))
}

func TestFormatStatementWarnings(t *testing.T) {
	warnings := []StatementWarning{
		{
			Severity:  "CRITICAL",
			Reason:    "HIGH_STATE_OPERATOR_WITHOUT_TTL",
			Message:   "Your query includes one or more highly state-intensive operators.",
			CreatedAt: ptrTime(time.Date(2026, 7, 30, 9, 15, 0, 0, time.UTC)),
		},
		{
			Severity:  "MODERATE",
			Reason:    "MISSING_WINDOW_START_END",
			Message:   "The GROUP BY clause contains only `window_start`.",
			CreatedAt: ptrTime(time.Date(2026, 7, 30, 8, 0, 0, 0, time.UTC)),
		},
	}

	expected := `Warnings:

CRITICAL [HIGH_STATE_OPERATOR_WITHOUT_TTL] (Logged: 2026-07-30T09:15:00Z)
Your query includes one or more highly state-intensive operators.

MODERATE [MISSING_WINDOW_START_END] (Logged: 2026-07-30T08:00:00Z)
The GROUP BY clause contains only ` + "`window_start`" + `.`

	require.Equal(t, expected, FormatStatementWarnings(warnings))
}

func TestFormatStatementWarningsSortsMostSevereFirstWithoutMutatingTheInput(t *testing.T) {
	warnings := []StatementWarning{
		{Severity: "LOW", Reason: "LOW_ONE", Message: "Low."},
		{Severity: "CRITICAL", Reason: "CRITICAL_ONE", Message: "Critical."},
	}

	formatted := FormatStatementWarnings(warnings)

	require.Less(t, strings.Index(formatted, "CRITICAL_ONE"), strings.Index(formatted, "LOW_ONE"))
	require.Equal(t, "LOW_ONE", warnings[0].Reason)
}

func TestFormatStatementWarningsOmitsMissingHeaderParts(t *testing.T) {
	warnings := []StatementWarning{{Severity: "LOW", Message: "Something to know."}}

	require.Equal(t, "Warnings:\n\nLOW\nSomething to know.", FormatStatementWarnings(warnings))
}

func TestFormatStatementWarningsFallsBackWhenSeverityIsMissing(t *testing.T) {
	warnings := []StatementWarning{{Reason: "SOME_REASON", Message: "Something to know."}}

	require.Equal(t, "Warnings:\n\nWARNING [SOME_REASON]\nSomething to know.", FormatStatementWarnings(warnings))
}

func TestFormatStatementWarningsRendersTimestampInUtc(t *testing.T) {
	berlin, err := time.LoadLocation("Europe/Berlin")
	require.NoError(t, err)

	warnings := []StatementWarning{{
		Severity:  "LOW",
		Reason:    "SOME_REASON",
		Message:   "Something to know.",
		CreatedAt: ptrTime(time.Date(2026, 7, 30, 11, 15, 0, 0, berlin)),
	}}

	require.Contains(t, FormatStatementWarnings(warnings), "(Logged: 2026-07-30T09:15:00Z)")
}
