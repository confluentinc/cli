package autocomplete

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/c-bata/go-prompt"
)

func TestSelectExampleAutoCompletionSnapshot(t *testing.T) {
	input := "select"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}

func TestDropExampleAutoCompletionSnapshot(t *testing.T) {
	input := "drop"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}

func TestAlterExampleAutoCompletionSnapshot(t *testing.T) {
	input := "alter"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}

func TestInserttExampleAutoCompletionSnapshot(t *testing.T) {
	input := "insert"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}

func TestDescribeExampleAutoCompletionSnapshot(t *testing.T) {
	input := "describe"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}

func TestExplainExampleAutoCompletionSnapshot(t *testing.T) {
	input := "explain"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}

func TestUseExampleAutoCompletionSnapshot(t *testing.T) {
	input := "use"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}

func TestShowExampleAutoCompletionSnapshot(t *testing.T) {
	input := "show"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}

func TestSetExampleAutoCompletionSnapshot(t *testing.T) {
	input := "set"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}

func TestResetExampleAutoCompletionSnapshot(t *testing.T) {
	input := "use"
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	actual := Completer(*buffer.Document())

	cupaloy.SnapshotT(t, actual)
}
