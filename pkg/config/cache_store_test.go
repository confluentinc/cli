package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCacheStore_RoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s := newCacheStore()

	require.NoError(t, s.writeJSON("probe.json", map[string]int{"n": 7}))

	var got map[string]int
	ok := s.readJSON("probe.json", &got)
	require.True(t, ok)
	require.Equal(t, 7, got["n"])
}

func TestCacheStore_MissingIsNotAnError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s := newCacheStore()

	var got map[string]int
	ok := s.readJSON("absent.json", &got)

	require.False(t, ok)
	require.Nil(t, got)
}

func TestLastUpdateCheckAt_PersistsToCacheNotConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	now := time.Now().UTC().Truncate(time.Second)

	c := New()
	c.LastUpdateCheckAt = &now
	require.NoError(t, c.Save())

	raw, err := os.ReadFile(c.GetFilename())
	require.NoError(t, err)
	require.NotContains(t, string(raw), "last_update_check_at")

	reloaded := New()
	reloaded.Filename = c.GetFilename()
	require.NoError(t, reloaded.Load())
	require.NotNil(t, reloaded.LastUpdateCheckAt)
	require.Equal(t, now, reloaded.LastUpdateCheckAt.UTC())
}
