package autocomplete

import (
	"github.com/confluentinc/go-prompt"
	"strings"
	"testing"

	"github.com/confluentinc/flink-sql-client/test/testutils"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func mockGetSmartCompletion() bool {
	return true
}

func TestBasicSelectAutoCompletion(t *testing.T) {
	input := "select"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	expected := prompt.Suggest{Text: "SELECT * FROM Orders WHERE amount = 2;", Description: "Select data from a database"}
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
	//buffer.CursorRight(2)

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
This could be remove if we eventually handle this formatting problems over at go-prompt
More about this: https://confluentinc.atlassian.net/jira/software/projects/KFS/boards/691?selectedIssue=KFS-606
*/
func TestNoLineBreaksInAutocompletion(t *testing.T) {
	statementGenerator := testutils.RandomStatementGenerator(3)

	// given
	rapid.Check(t, func(t *rapid.T) {
		randomStatement := statementGenerator.Example()
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
