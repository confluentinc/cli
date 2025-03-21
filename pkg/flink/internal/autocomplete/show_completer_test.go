package autocomplete

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/confluentinc/go-prompt"
)

func TestShowAutoCompletionSnapshot(t *testing.T) {
	input := "show"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	completer := NewCompleterBuilder(mockCompletionsEnabled).
		AddCompleter(ShowCompleter).
		BuildCompleter()

	actual := completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}
