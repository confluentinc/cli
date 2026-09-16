package config

import (
	"os"
	"path/filepath"
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

// newTestConfigWithOneContext builds a valid Config with one context named "ctx",
// mirroring config_concurrent_test.go's createContextReusingAPIKey helper.
func newTestConfigWithOneContext(t *testing.T) *Config {
	t.Helper()
	c := New()
	c.Filename = filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, c.Load())
	require.NoError(t, c.CreateContext("ctx", "https://example.com", "test", "api-secret-value"))
	return c
}

func TestFeatureFlags_PersistToCacheNotContexts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
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
