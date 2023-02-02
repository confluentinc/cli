package components

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/c-bata/go-prompt"
)

func TestSelectExampleAutoCompletionSnapshot(t *testing.T) {
	input := "select"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}
