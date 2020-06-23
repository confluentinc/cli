package errors

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuggestionsMessage(t *testing.T) {
	errorMessage := "im an error hi"
	suggestionsMessage := "This is a suggestion"
	err := NewErrorWithSuggestions(errorMessage, suggestionsMessage)
	var b bytes.Buffer
	HandleSuggestionsMessageDisplay(err, &b)
	out := b.String()
	wantDirectionsOutput := fmt.Sprintf(suggestionsMessageFormat, suggestionsMessage)
	require.Equal(t, wantDirectionsOutput, out)
}

