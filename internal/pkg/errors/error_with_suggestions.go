package errors

import (
	"fmt"
	"io"
	"strings"
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

func DisplaySuggestionsMessage(err error, writer io.Writer) {
	if err == nil {
		return
	}
	cliErr, ok := err.(ErrorWithSuggestions)
	if ok && cliErr.GetSuggestionsMsg() != "" {
		_, _ = fmt.Fprint(writer, ComposeSuggestionsMessage(cliErr.GetSuggestionsMsg()))
	}
}

func ComposeSuggestionsMessage(msg string) string {
	lines := strings.Split(msg, "\n")
	suggestionsMsg := suggestionsMessageHeader
	for _, line := range lines {
		suggestionsMsg += fmt.Sprintf(suggestionsLineFormat, line)
	}
	return suggestionsMsg
}
