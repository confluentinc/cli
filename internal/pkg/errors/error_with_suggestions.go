package errors

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
