package flink

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/require"
)

func TestReplaceFile(t *testing.T) {
	// requireFile asserts path's content and permission bits, and that no temporary file was left beside it. Windows
	// only tracks the read-only attribute, so the permission check is skipped there.
	requireFile := func(t *testing.T, path, content string, perm os.FileMode) {
		actual, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, content, string(actual))

		if runtime.GOOS != "windows" {
			info, err := os.Stat(path)
			require.NoError(t, err)
			require.Equal(t, perm, info.Mode().Perm())
		}

		entries, err := os.ReadDir(filepath.Dir(path))
		require.NoError(t, err)
		require.Len(t, entries, 1, "the temporary file must be removed")
	}

	t.Run("creates a new file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "artifact.jar")

		require.NoError(t, replaceFile(path, strings.NewReader("v1 content")))

		requireFile(t, path, "v1 content", 0644)
	})

	t.Run("replaces an existing file and keeps its permissions", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "artifact.jar")
		require.NoError(t, os.WriteFile(path, []byte("a longer previous version"), 0600))
		require.NoError(t, os.Chmod(path, 0600))

		require.NoError(t, replaceFile(path, strings.NewReader("v2")))

		requireFile(t, path, "v2", 0600)
	})

	t.Run("leaves an existing file untouched when the write fails", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "artifact.jar")
		require.NoError(t, os.WriteFile(path, []byte("valid artifact"), 0644))
		require.NoError(t, os.Chmod(path, 0644))

		failing := io.MultiReader(strings.NewReader("partial"), iotest.ErrReader(errors.New("disk full")))
		require.ErrorContains(t, replaceFile(path, failing), "failed to write output file")

		requireFile(t, path, "valid artifact", 0644)
	})
}
