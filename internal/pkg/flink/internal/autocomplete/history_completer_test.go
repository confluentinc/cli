package autocomplete

import (
	"testing"


	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/go-prompt"
)

func TestHistorySnapshot(t *testing.T) {
	input := "select sp"
	history := []string{"select spec from table1", "select spec2 from table1"}
	buffer := prompt.NewBuffer()
	buffer.InsertText(input, false, true)

	completer := NewCompleterBuilder(mockGetSmartCompletion).
		AddCompleter(GenerateHistoryCompleter(history)).
		BuildCompleter()

	suggestions := completer(*buffer.Document())

	cupaloy.SnapshotT(t, suggestions)
	require.Len(t, suggestions, 2)
	require.Equal(t, "spec2 from", suggestions[0].Text)
	require.Equal(t, "spec from", suggestions[1].Text)
}
