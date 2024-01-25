package flink

import (
	"net/http"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

// Error extends the ErrorWithSuggestion with a status code.
type Error struct {
	errorMsg       string
	suggestionsMsg string
	statusCode     int
}

func NewError(errorMsg string, suggestionsMsg string, statusCode int) Error {
	return Error{
		errorMsg:       errorMsg,
		suggestionsMsg: suggestionsMsg,
		statusCode:     statusCode,
	}
}

func (e Error) StatusCode() int {
	return e.statusCode
}

func (e Error) GetSuggestionsMsg() string {
	return e.suggestionsMsg
}

func (e Error) Error() string {
	return e.errorMsg
}

type Coder interface {
	StatusCode() int
}

var _ Coder = (*Error)(nil)
var _ errors.ErrorWithSuggestions = (*Error)(nil)

// Extends error with status code, including suggestion if err type is ErrorWithSuggestion
func CatchError(err error, r *http.Response) error {
	if err == nil {
		return nil
	}
	err = errors.CatchCCloudV2Error(err, r)
	suggestionsMsg := ""
	if suggester, ok := err.(errors.ErrorWithSuggestions); ok {
		suggestionsMsg = suggester.GetSuggestionsMsg()
	}
	var statusCode int
	if r != nil {
		statusCode = r.StatusCode
	}
	return Error{
		statusCode:     statusCode,
		errorMsg:       err.Error(),
		suggestionsMsg: suggestionsMsg,
	}
}
