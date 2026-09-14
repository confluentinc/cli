package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteFileAtomic_ReplacesExistingAndSetsPerms(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte("old"), 0600))

	require.NoError(t, writeFileAtomic(path, []byte("new-contents")))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "new-contents", string(got))

	info, err := os.Stat(path)
	require.NoError(t, err)
	// Windows does not expose POSIX permission bits through Mode().Perm(), so the
	// 0600 assertion only holds off Windows (matching config_test.go's own guard).
	if runtime.GOOS != "windows" {
		require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	}
}

// A symlinked config path must be followed like os.WriteFile did on main: the real
// target is replaced atomically and the symlink is preserved, not swapped for a
// regular file. Skipped on Windows, where symlinks need privilege and differ.
func TestWriteFileAtomic_FollowsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics and privileges differ on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "real.json")
	link := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(target, []byte("old"), 0600))
	require.NoError(t, os.Symlink(target, link))

	require.NoError(t, writeFileAtomic(link, []byte("new-contents")))

	info, err := os.Lstat(link)
	require.NoError(t, err)
	require.NotZero(t, info.Mode()&os.ModeSymlink, "config.json must stay a symlink, not become a regular file")

	got, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "new-contents", string(got), "the symlink's target must receive the new contents")
}

func TestWriteFileAtomic_LeavesNoTempOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	require.NoError(t, writeFileAtomic(path, []byte("x")))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "temp file was not cleaned up / renamed")
	require.Equal(t, "config.json", entries[0].Name())
}
