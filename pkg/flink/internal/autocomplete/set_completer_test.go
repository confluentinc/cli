package autocomplete

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/confluentinc/go-prompt"
)

func TestSetAutoCompletionSnapshot(t *testing.T) {
	input := "set"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	completer := NewCompleterBuilder(mockCompletionsEnabled).
		AddCompleter(SetCompleterCommon).
		AddCompleter(SetCompleterCloud).
		BuildCompleter()

	actual := completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}
