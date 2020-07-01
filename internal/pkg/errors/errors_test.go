package errors

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	wantSuggestionsMsgFormat = `
Suggestions:
    %s
`
)

func TestSuggestionsMessage(t *testing.T) {
	errorMessage := "im an error hi"
	suggestionsMessage := "This is a suggestion"
	err := NewErrorWithSuggestions(errorMessage, suggestionsMessage)
	var b bytes.Buffer
	DisplaySuggestionsMessage(err, &b)
	out := b.String()
	wantSuggestionsMsg := fmt.Sprintf(wantSuggestionsMsgFormat, suggestionsMessage)
	require.Equal(t, wantSuggestionsMsg, out)
}
