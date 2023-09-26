package autocomplete

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"

	"github.com/confluentinc/go-prompt"
)

func TestSetAutoCompletionSnapshot(t *testing.T) {
	input := "set"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	completer := NewCompleterBuilder(mockGetSmartCompletion).
		AddCompleter(SetCompleter).
		BuildCompleter()

	actual := completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}
