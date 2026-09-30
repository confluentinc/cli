package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddAndRemove(t *testing.T) {
	expectedWarnings := []string{
		`"a" is marked for addition and deletion`,
		`"b" is marked for addition but already exists`,
		`"c" is marked for deletion but does not exist`,
	}

	result, warnings := AddAndRemove([]string{"b"}, []string{"a", "b"}, []string{"a", "c"})
	require.ElementsMatch(t, result, []string{"b"})
	require.Equal(t, expectedWarnings, warnings)
}
