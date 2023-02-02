package components

import (
	"testing"

	"github.com/c-bata/go-prompt"
)

func containsSuggestion(suggestions []prompt.Suggest, s prompt.Suggest) bool {
	for _, v := range suggestions {
		if v.Text == s.Text {
			return true
		}
	}

	return false
}

func TestBasicSelectAutoCompletion(t *testing.T) {
	input := "select"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	expected := prompt.Suggest{Text: "SELECT * FROM Orders WHERE amount = 2;", Description: "Select data from a database"}
	actual := completer(*buffer.Document())

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
	actual := completer(*buffer.Document())

	if containsSuggestion(actual, expected) {
		t.Errorf("prompt.Run() = %q, want %q", actual, expected)
	}
}

func TestAutoCompletionWithHistory(t *testing.T) {
	input := "select"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	expected := prompt.Suggest{Text: "SELECT * FROM YESTERDAY;", Description: "History entry"}
	completerWithHistory := completerWithHistory([]string{"SELECT * FROM YESTERDAY;"})
	actual := completerWithHistory(*buffer.Document())

	if !containsSuggestion(actual, expected) {
		t.Errorf("prompt.Run() = %q, want %q", actual, expected)
	}
}
