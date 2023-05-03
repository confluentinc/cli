package autocomplete

import (
	"github.com/confluentinc/go-prompt"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestShowAutoCompletionSnapshot(t *testing.T) {
	input := "show"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	completer := NewCompleterBuilder(mockGetSmartCompletion).
		AddCompleter(ShowCompleter).
		BuildCompleter()

	actual := completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}
