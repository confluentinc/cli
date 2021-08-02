package examples

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildExampleString(t *testing.T) {
	got := BuildExampleString(
		Example{
			Text: "Text",
			Code: "Code",
		},
		Example{
			Text: "Text",
			Code: "confluent",
		},
	)

	want := "Text\n\n  Code\n\nText\n\n  $ confluent"
	require.Equal(t, want, got)
}
