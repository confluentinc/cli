package errors

type ErrorWithSuggestions interface {
	error
	GetSuggestionsMsg() string
}

type ErrorWithSuggestionsImpl struct {
	errorMessage string
	suggestionsMessage string
}

func NewErrorWithSuggestions(errorMessage string, suggestionsMessage string) ErrorWithSuggestions {
	return &ErrorWithSuggestionsImpl{
		errorMessage:       errorMessage,
		suggestionsMessage: suggestionsMessage,
	}
}

func (b *ErrorWithSuggestionsImpl) Error() string {
	return b.errorMessage
}

func (b *ErrorWithSuggestionsImpl) GetSuggestionsMsg() string {
	return b.suggestionsMessage
}
