package autocomplete

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
)

func mockGetSmartCompletion() bool {
	return true
}

func TestBasicSelectAutoCompletion(t *testing.T) {
	input := "select"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	expected := prompt.Suggest{Text: "SELECT ", Description: "Select data from a database"}
	completer := NewCompleterBuilder(mockGetSmartCompletion).
		AddCompleter(ExamplesCompleter).
		AddCompleter(SetCompleter).
		AddCompleter(ShowCompleter).
		BuildCompleter()

	suggestions := completer(*buffer.Document())
	require.Contains(t, suggestions, expected)
}

func TestFailingBasicAutoCompletion(t *testing.T) {
	input := "non-existing-statement"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	expected := prompt.Suggest{Text: "SELECT", Description: "Select data from a database"}
	completer := NewCompleterBuilder(mockGetSmartCompletion).
		AddCompleter(ExamplesCompleter).
		AddCompleter(SetCompleter).
		AddCompleter(ShowCompleter).
		BuildCompleter()

	suggestions := completer(*buffer.Document())
	require.NotContains(t, suggestions, expected)
}

/*
Line breaks in suggestions breaks go-prompt render/formatting when you go through suggestions.
Apart from that, they're not properly formatted for visualization.
This test makes sure that we don't generate suggestions with line breaks.
This could be removed if we eventually handle this formatting problem over at go-prompt
*/
func TestNoLineBreaksInAutocompletion(t *testing.T) {
	// given
	rapid.Check(t, func(t *rapid.T) {
		randomStatement := generators.RandomSQLSentence().Example()
		buffer := prompt.NewBuffer()

		// when
		buffer.InsertText(randomStatement.Text, false, true)
		completer := NewCompleterBuilder(mockGetSmartCompletion).
			AddCompleter(ExamplesCompleter).
			AddCompleter(SetCompleter).
			AddCompleter(ShowCompleter).
			BuildCompleter()
		suggestions := completer(*buffer.Document())

		// then
		for _, element := range suggestions {
			require.Equalf(t, strings.Contains(element.Text, "\n"), false, "Suggestions are not allowed to have line breaks. The following suggestion caused the error:\n %s", element.Text)
		}
	})
}
