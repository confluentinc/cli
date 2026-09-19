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

// A multi-hop symlink chain ending at a not-yet-existing target must be followed the
// whole way, like os.WriteFile did: the chain's links are preserved and the final
// target is created. Skipped on Windows, where symlinks need privilege and differ.
func TestWriteFileAtomic_FollowsDanglingSymlinkChain(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics and privileges differ on Windows")
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "real.json") // does not exist yet
	mid := filepath.Join(dir, "b")            // b -> real.json (dangling)
	link := filepath.Join(dir, "config.json") // config.json -> b
	require.NoError(t, os.Symlink(target, mid))
	require.NoError(t, os.Symlink(mid, link))

	require.NoError(t, writeFileAtomic(link, []byte("new-contents")))

	linkInfo, err := os.Lstat(link)
	require.NoError(t, err)
	require.NotZero(t, linkInfo.Mode()&os.ModeSymlink, "config.json must stay a symlink")
	midInfo, err := os.Lstat(mid)
	require.NoError(t, err)
	require.NotZero(t, midInfo.Mode()&os.ModeSymlink, "the intermediate link must stay a symlink")

	got, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "new-contents", string(got), "the chain's final target must be created with the contents")
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
