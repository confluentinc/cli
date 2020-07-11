package examples

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildExampleString(t *testing.T) {
	got := BuildExampleString(
		Example{
			Desc: "Desc",
			Code: "Code",
		},
	)

	want := "Desc\n\n::\n\n  Code\n\n"
	require.Equal(t, got, want)
}

func TestTab(t *testing.T) {
	require.Equal(t, tab("A\nB"), "  A\n  B")
}
