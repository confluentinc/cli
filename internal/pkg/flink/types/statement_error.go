package types

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type StatementError struct {
	Message        string
	StatusCode     int
	FailureMessage string
	Usage          []string
	Suggestion     string
}

func NewStatementError(err error) *StatementError {
	return &StatementError{
		Message:    err.Error(),
		StatusCode: StatusCode(err),
	}
}

func NewStatementErrorFailureMsg(err error, failureMsg string) *StatementError {
	return &StatementError{
		Message:        err.Error(),
		StatusCode:     StatusCode(err),
		FailureMessage: failureMsg,
	}
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

// StatusCode extract the status code if the error implements Coder interface.
func StatusCode(err error) int {
	if coder, ok := err.(ccloudv2.Coder); ok {
		return coder.StatusCode()
	}
	return 0
}
