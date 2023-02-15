package pipeline

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPath_NoPrefix(t *testing.T) {
	require.Equal(t, "pipeline.sql", getPath("pipeline.sql"))
}

func TestGetPath_HomeDir(t *testing.T) {
	err := os.Setenv("HOME", "home")
	require.NoError(t, err)

	require.Equal(t, "home/pipeline.sql", getPath("~/pipeline.sql"))
}

func TestGetPath_HomeDirUnset(t *testing.T) {
	err := os.Unsetenv("HOME")
	require.NoError(t, err)

	require.Equal(t, "~/pipeline.sql", getPath("~/pipeline.sql"))
}
