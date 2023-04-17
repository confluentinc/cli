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
	suggestionsMsg string
}

func NewErrorWithSuggestions(errorMsg string, suggestionsMsg string) ErrorWithSuggestions {
	return &ErrorWithSuggestionsImpl{
		errorMsg:       errorMsg,
		suggestionsMsg: suggestionsMsg,
	}
}

func NewWrapErrorWithSuggestions(err error, errorMsg string, suggestionsMsg string) ErrorWithSuggestions {
	return &ErrorWithSuggestionsImpl{
		errorMsg:       Wrap(err, errorMsg).Error(),
		suggestionsMsg: suggestionsMsg,
	}
}

func (b *ErrorWithSuggestionsImpl) Error() string {
	return b.errorMsg
}

func (b *ErrorWithSuggestionsImpl) GetSuggestionsMsg() string {
	return b.suggestionsMsg
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
