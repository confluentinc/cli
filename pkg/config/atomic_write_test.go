package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteFileAtomic_ReplacesExistingAndSetsPerms(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte("old"), 0600))

	require.NoError(t, writeFileAtomic(path, []byte("new-contents"), 0600))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "new-contents", string(got))

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

func TestWriteFileAtomic_LeavesNoTempOnSuccess(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	require.NoError(t, writeFileAtomic(path, []byte("x"), 0600))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "temp file was not cleaned up / renamed")
	require.Equal(t, "config.json", entries[0].Name())
}
