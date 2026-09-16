package config

import (
	"testing"

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
