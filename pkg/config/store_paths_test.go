package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheDir_UnderStateDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := CacheDir()

	require.Equal(t, filepath.Join(home, StateDirName(), ".cache"), got)
}
