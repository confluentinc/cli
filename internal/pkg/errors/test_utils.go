package errors

import (
	"bytes"
	"fmt"

	"github.com/stretchr/testify/require"
)

var (
	errorAndSuggestionsMessageFormat = "Error: %s\n%s"
)

func VerifyErrorAndSuggestions(assertions *require.Assertions, err error, expectedErrorMsg string, expectedSuggestions string) {
	assertions.Equal(expectedErrorOutput(expectedErrorMsg, expectedSuggestions), GetErrorStringWithSuggestions(err))
}

func GetErrorStringWithSuggestions(err error) string {
	var b bytes.Buffer
	DisplaySuggestionsMessage(err, &b)
	out := b.String()
	if out == "" {
		return err.Error()
	}
	return "Error: " + err.Error() + "\n" + out
}

func expectedErrorOutput(errMsg string, suggestionsMsg string) string {
	suggestionsMsg = ComposeSuggestionsMessage(suggestionsMsg)
	return fmt.Sprintf(errorAndSuggestionsMessageFormat, errMsg, suggestionsMsg)
}
