package pipeline

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPath_NoPrefix(t *testing.T) {
	require.Equal(t, "pipeline.sql", expandHomeDir("pipeline.sql"))
}

func TestGetPath_HomeDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", "userprofile")
		require.Equal(t, `userprofile\pipeline.sql`, expandHomeDir(`~\pipeline.sql`))
	} else {
		t.Setenv("HOME", "home")
		require.Equal(t, "home/pipeline.sql", expandHomeDir("~/pipeline.sql"))
	}
}

func TestGetPath_HomeDirUnset(t *testing.T) {
	if runtime.GOOS == "windows" {
		err := os.Unsetenv("USERPROFILE")
		require.NoError(t, err)
		require.Equal(t, `~\pipeline.sql`, expandHomeDir(`~\pipeline.sql`))
	} else {
		err := os.Unsetenv("HOME")
		require.NoError(t, err)
		require.Equal(t, "~/pipeline.sql", expandHomeDir("~/pipeline.sql"))
	}
}
