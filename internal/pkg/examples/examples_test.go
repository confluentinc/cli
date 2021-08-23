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
			Code: "cflt",
		},
	)

	want := "Text\n\n  Code\n\nText\n\n  $ cflt"
	require.Equal(t, want, got)
}
