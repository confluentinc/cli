package types

import (
	"fmt"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

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
		errStr += fmt.Sprintf("\nUsage: %s", utils.ArrayToCommaDelimitedString(e.Usage, "or"))
	}
	if e.Suggestion != "" {
		errStr += fmt.Sprintf("\nSuggestion: %s", e.Suggestion)
	}
	if e.FailureMessage != "" {
		errStr += fmt.Sprintf("\nError details: %s", e.FailureMessage)
	}

	return errStr
}
