package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCacheStore_RoundTrip(t *testing.T) {
	setTestHome(t, t.TempDir())
	s := newCacheStore()

	require.NoError(t, s.writeJSON("probe.json", map[string]int{"n": 7}))

	var got map[string]int
	ok := s.readJSON("probe.json", &got)
	require.True(t, ok)
	require.Equal(t, 7, got["n"])
}

func TestCacheStore_MissingIsNotAnError(t *testing.T) {
	setTestHome(t, t.TempDir())
	s := newCacheStore()

	var got map[string]int
	ok := s.readJSON("absent.json", &got)

	require.False(t, ok)
	require.Nil(t, got)
}

func TestLastUpdateCheckAt_PersistsToCacheNotConfig(t *testing.T) {
	setTestHome(t, t.TempDir())
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

// setTestHome points os.UserHomeDir at dir on both Unix (HOME) and Windows
// (USERPROFILE). A test that sets only HOME writes to the real profile on the
// Windows CI runner, where os.UserHomeDir consults USERPROFILE instead.
func setTestHome(t *testing.T, dir string) {
	t.Helper()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
}

// newTestConfigWithOneContext builds a valid Config with one context named "ctx".
func newTestConfigWithOneContext(t *testing.T) *Config {
	t.Helper()
	c := New()
	c.Filename = filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, c.Load())
	require.NoError(t, c.CreateContext("ctx", "https://example.com", "test", "api-secret-value"))
	return c
}

func TestFeatureFlags_PersistToCacheNotContexts(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithOneContext(t)
	c.Contexts["ctx"].FeatureFlags = &FeatureFlags{CliValues: map[string]any{"flag": true}}
	require.NoError(t, c.Save())

	raw, err := os.ReadFile(c.GetFilename())
	require.NoError(t, err)
	// "ccloud_values" is unique to the FeatureFlags struct: a bare "feature_flags"
	// check would false-positive on the unrelated top-level "disable_feature_flags" field.
	require.NotContains(t, string(raw), "ccloud_values")

	reloaded := New()
	reloaded.Filename = c.GetFilename()
	require.NoError(t, reloaded.Load())
	require.Equal(t, true, reloaded.Contexts["ctx"].FeatureFlags.CliValues["flag"])
}

func TestLoad_MigrationDoesNotClobberCachedTimestamp(t *testing.T) {
	setTestHome(t, t.TempDir())
	now := time.Now().UTC().Truncate(time.Second)

	// Seed a config whose deprecated flag forces a migration-driven Save() on the
	// next Load, then cache an update-check timestamp the migration must preserve.
	seed := New()
	seed.DisablePluginsOnce = true
	require.NoError(t, seed.Save())
	require.NoError(t, newCacheStore().writeJSON("update_check.json", updateCheckCache{LastUpdateCheckAt: &now}))

	reloaded := New()
	reloaded.Filename = seed.GetFilename()
	require.NoError(t, reloaded.Load())

	require.NotNil(t, reloaded.LastUpdateCheckAt)
	require.Equal(t, now, reloaded.LastUpdateCheckAt.UTC())
}

func TestFeatureFlagCache_ClearedContextDropsStaleEntry(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithOneContext(t)
	c.Contexts["ctx"].FeatureFlags = &FeatureFlags{CliValues: map[string]any{"flag": true}}
	require.NoError(t, c.Save())

	// Clearing a context's flags must drop its cache entry; otherwise a same-named
	// context created later resurrects the stale flags and suppresses a refetch.
	c.Contexts["ctx"].FeatureFlags = nil
	require.NoError(t, c.Save())

	flags := map[string]*FeatureFlags{}
	require.True(t, newCacheStore().readJSON("feature_flags.json", &flags))
	require.NotContains(t, flags, "ctx")
}

func TestSave_CacheWriteFailureDoesNotFailSave(t *testing.T) {
	home := t.TempDir()
	setTestHome(t, home)
	stateDir := filepath.Join(home, StateDirName())
	require.NoError(t, os.MkdirAll(stateDir, 0700))
	// A regular file where the cache dir should be makes os.MkdirAll(.cache) fail.
	require.NoError(t, os.WriteFile(filepath.Join(stateDir, ".cache"), []byte("x"), 0600))

	c := New()
	now := time.Now()
	c.LastUpdateCheckAt = &now

	require.NoError(t, c.Save()) // config write must still succeed
	require.FileExists(t, c.GetFilename())
}
