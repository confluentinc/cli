package flink

import (
	"net/http"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

// FlinkError extends the ErrorWithSuggestion with a status code.
type FlinkError struct {
	errorMsg       string
	suggestionsMsg string
	statusCode     int
}

func NewFlinkError(errorMsg string, suggestionsMsg string, statusCode int) FlinkError {
	return FlinkError{
		errorMsg:       errorMsg,
		suggestionsMsg: suggestionsMsg,
		statusCode:     statusCode,
	}
}

func (f FlinkError) StatusCode() int {
	return f.statusCode
}

func (f FlinkError) GetSuggestionsMsg() string {
	return f.suggestionsMsg
}

func (f FlinkError) Error() string {
	return f.errorMsg
}

type Coder interface {
	StatusCode() int
}

var _ Coder = (*FlinkError)(nil)
var _ errors.ErrorWithSuggestions = (*FlinkError)(nil)

// Extends error with status code, including suggestion if err type is ErrorWithSuggestion
func CatchFlinkError(err error, r *http.Response) error {
	if err == nil {
		return nil
	}
	err = errors.CatchCCloudV2Error(err, r)
	suggestion := ""
	if suggester, ok := err.(errors.ErrorWithSuggestions); ok {
		suggestion = suggester.GetSuggestionsMsg()
	}
	var statusCode int
	if r != nil {
		statusCode = r.StatusCode
	}
	return FlinkError{
		statusCode:     statusCode,
		errorMsg:       err.Error(),
		suggestionsMsg: suggestion,
	}
}
