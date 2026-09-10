package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// helper: a config with one platform keyed by name
func cfgWithPlatforms(names ...string) *Config {
	c := New()
	for _, n := range names {
		c.Platforms[n] = &Platform{Name: n}
	}
	return c
}

func TestThreeWayMerge_PreservesConcurrentAdd(t *testing.T) {
	base := cfgWithPlatforms("a")
	ours := cfgWithPlatforms("a")      // we changed nothing about platforms
	disk := cfgWithPlatforms("a", "b") // another session added "b"

	got := threeWayMerge(base, ours, disk)

	require.Contains(t, got.Platforms, "a")
	require.Contains(t, got.Platforms, "b", "a concurrent add on disk must survive the merge")
}

func TestThreeWayMerge_AppliesOurDelete(t *testing.T) {
	base := cfgWithPlatforms("a", "b")
	ours := cfgWithPlatforms("a")      // we deleted "b"
	disk := cfgWithPlatforms("a", "b") // disk still has "b"

	got := threeWayMerge(base, ours, disk)

	require.Contains(t, got.Platforms, "a")
	require.NotContains(t, got.Platforms, "b", "our delete must be applied to disk state")
}

func TestThreeWayMerge_OurScalarChangeWins(t *testing.T) {
	base := New()
	base.CurrentContext = "a"
	ours := New()
	ours.CurrentContext = "b" // we ran `use b`
	disk := New()
	disk.CurrentContext = "a" // disk unchanged

	got := threeWayMerge(base, ours, disk)

	require.Equal(t, "b", got.CurrentContext)
}

func TestThreeWayMerge_UntouchedScalarTakesDisk(t *testing.T) {
	base := New()
	base.EnableColor = false
	ours := New()
	ours.EnableColor = false // we didn't touch it
	disk := New()
	disk.EnableColor = true // another session enabled color

	got := threeWayMerge(base, ours, disk)

	require.True(t, got.EnableColor, "a field we didn't change must keep the disk value")
}
