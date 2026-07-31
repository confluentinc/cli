package types

import (
	"fmt"
	"slices"
	"strings"
	"time"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
)

// severityOrder ranks severities for display, most severe first. Severity is an extensible enum, so
// an unrecognized value is still displayed, sorted last.
var severityOrder = []string{"CRITICAL", "MODERATE", "LOW"}

// StatementWarning is a non-fatal issue reported for a statement.
type StatementWarning struct {
	Severity string `json:"severity" yaml:"severity"`
	Reason   string `json:"reason" yaml:"reason"`
	Message  string `json:"message" yaml:"message"`
	// A pointer so that a missing timestamp is omitted from serialized output. `omitempty` does not
	// omit a zero `time.Time` value, only a nil pointer.
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"created_at,omitempty"`
}

// NewStatementWarnings converts a statement's warnings, ordered most severe first.
func NewStatementWarnings(warnings []flinkgatewayv1.SqlV1StatementWarning) []StatementWarning {
	if len(warnings) == 0 {
		return nil
	}

	converted := make([]StatementWarning, len(warnings))
	for i, warning := range warnings {
		converted[i] = StatementWarning{
			Severity: warning.GetSeverity(),
			Reason:   warning.GetReason(),
			Message:  warning.GetMessage(),
		}
		// The API models the timestamp as a value, so an absent one arrives as the zero time.
		if createdAt := warning.GetCreatedAt(); !createdAt.IsZero() {
			converted[i].CreatedAt = &createdAt
		}
	}

	sortBySeverity(converted)

	return converted
}

// FormatStatementWarnings renders warnings for terminal output, most severe first. It returns an
// empty string when there are no warnings.
func FormatStatementWarnings(warnings []StatementWarning) string {
	if len(warnings) == 0 {
		return ""
	}

	// Sort a copy so the order holds however the caller built the slice.
	sorted := slices.Clone(warnings)
	sortBySeverity(sorted)

	entries := make([]string, len(sorted))
	for i, warning := range sorted {
		entries[i] = fmt.Sprintf("%s\n%s", warning.header(), warning.Message)
	}

	return fmt.Sprintf("Warnings:\n\n%s", strings.Join(entries, "\n\n"))
}

func sortBySeverity(warnings []StatementWarning) {
	slices.SortStableFunc(warnings, func(a, b StatementWarning) int {
		return severityRank(a.Severity) - severityRank(b.Severity)
	})
}

func (w StatementWarning) header() string {
	severity := w.Severity
	if severity == "" {
		severity = "WARNING"
	}

	header := severity
	if w.Reason != "" {
		header += fmt.Sprintf(" [%s]", w.Reason)
	}
	if w.CreatedAt != nil {
		header += fmt.Sprintf(" (Logged: %s)", w.CreatedAt.UTC().Format(time.RFC3339))
	}

	return header
}

func severityRank(severity string) int {
	if i := slices.Index(severityOrder, strings.ToUpper(severity)); i >= 0 {
		return i
	}
	return len(severityOrder)
}
