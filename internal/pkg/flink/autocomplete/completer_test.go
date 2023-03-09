package autocomplete

import (
	"strings"
	"testing"

	"github.com/c-bata/go-prompt"
	"github.com/confluentinc/flink-sql-client/test/testutils"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func containsSuggestion(suggestions []prompt.Suggest, s prompt.Suggest) bool {
	for _, v := range suggestions {
		if v.Text == s.Text {
			return true
		}
	}

	return false
}
func mockGetSmartCompletion() bool {
	return true
}

func TestBasicSelectAutoCompletion(t *testing.T) {
	input := "select"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	expected := prompt.Suggest{Text: "SELECT * FROM Orders WHERE amount = 2;", Description: "Select data from a database"}
	actual := Completer(*buffer.Document())

	if !containsSuggestion(actual, expected) {
		t.Errorf("prompt.Run() = %q, want %q", actual, expected)
	}
}

func TestFailingBasicAutoCompletion(t *testing.T) {
	input := "non-existing-statement"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)
	//buffer.CursorRight(2)

	expected := prompt.Suggest{Text: "SELECT", Description: "Select data from a database"}
	actual := Completer(*buffer.Document())

	if containsSuggestion(actual, expected) {
		t.Errorf("prompt.Run() = %q, want %q", actual, expected)
	}
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
		suggestions := Completer(*buffer.Document())

		// then
		for _, element := range suggestions {
			require.Equalf(t, strings.Contains(element.Text, "\n"), false, "Suggestions are not allowed to have line breaks. The following suggestion caused the error:\n %s", element.Text)
		}
	})
}
