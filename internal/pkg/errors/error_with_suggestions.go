package errors

import (
	"fmt"
	"strings"
)

var (
	suggestionsMessageHeader = "\nSuggestions:\n"
	suggestionsLineFormat    = "    %s\n"
)

type ErrorWithSuggestions interface {
	error
	GetSuggestionsMsg() string
}

type ErrorWithSuggestionsImpl struct {
	errorMsg       string
	statusCode     int
	suggestionsMsg string
}

func NewErrorWithSuggestions(errorMsg string, suggestionsMsg string) ErrorWithSuggestions {
	return &ErrorWithSuggestionsImpl{
		errorMsg:       errorMsg,
		suggestionsMsg: suggestionsMsg,
	}
}

func NewErrorWithSuggestionsAndCode(errorMsg string, suggestionsMsg string, statusCode int) ErrorWithSuggestions {
	return &ErrorWithSuggestionsImpl{
		errorMsg:       errorMsg,
		suggestionsMsg: suggestionsMsg,
		statusCode:     statusCode,
	}
}

func NewWrapErrorWithSuggestions(err error, errorMsg string, suggestionsMsg string) ErrorWithSuggestions {
	return &ErrorWithSuggestionsImpl{
		errorMsg:       Wrap(err, errorMsg).Error(),
		suggestionsMsg: suggestionsMsg,
	}
}

func NewWrapErrorWithSuggestionsAndCode(err error, errorMsg string, suggestionsMsg string, statusCode int) ErrorWithSuggestions {
	return &ErrorWithSuggestionsImpl{
		errorMsg:       Wrap(err, errorMsg).Error(),
		suggestionsMsg: suggestionsMsg,
		statusCode:     statusCode,
	}
}

func (b *ErrorWithSuggestionsImpl) Error() string {
	return b.errorMsg
}

func (b *ErrorWithSuggestionsImpl) GetSuggestionsMsg() string {
	return b.suggestionsMsg
}

func (b *ErrorWithSuggestionsImpl) GetStatusCode() int {
	return b.statusCode
}

func DisplaySuggestionsMessage(err error) string {
	if err, ok := err.(ErrorWithSuggestions); ok {
		if suggestion := err.GetSuggestionsMsg(); suggestion != "" {
			return ComposeSuggestionsMessage(suggestion)
		}
	}
	return ""
}

func ComposeSuggestionsMessage(msg string) string {
	lines := strings.Split(msg, "\n")
	suggestionsMsg := suggestionsMessageHeader
	for _, line := range lines {
		suggestionsMsg += fmt.Sprintf(suggestionsLineFormat, line)
	}
	return suggestionsMsg
}

type Coder interface {
	GetStatusCode() int
}

// StatusCode extract the status code if the error implements Coder interface.
func StatusCode(err error) int {
	if coder, ok := err.(Coder); ok {
		return coder.GetStatusCode()
	}
	return 0
}
