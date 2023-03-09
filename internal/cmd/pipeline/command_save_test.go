package pipeline

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPath_NoPrefix(t *testing.T) {
	t.Parallel()

	require.Equal(t, "pipeline.sql", getPath("pipeline.sql"))
}

func TestGetPath_HomeDir(t *testing.T) { //nolint:paralleltest
	t.Setenv("HOME", "home")
	require.Equal(t, "home/pipeline.sql", getPath("~/pipeline.sql"))
}

func TestGetPath_HomeDirUnset(t *testing.T) { //nolint:paralleltest
	err := os.Unsetenv("HOME")
	require.NoError(t, err)

	require.Equal(t, "~/pipeline.sql", getPath("~/pipeline.sql"))
}
