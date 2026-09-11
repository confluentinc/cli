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

	got, err := threeWayMerge(base, ours, disk)
	require.NoError(t, err)

	require.Contains(t, got.Platforms, "a")
	require.Contains(t, got.Platforms, "b", "a concurrent add on disk must survive the merge")
}

func TestThreeWayMerge_AppliesOurDelete(t *testing.T) {
	base := cfgWithPlatforms("a", "b")
	ours := cfgWithPlatforms("a")      // we deleted "b"
	disk := cfgWithPlatforms("a", "b") // disk still has "b"

	got, err := threeWayMerge(base, ours, disk)
	require.NoError(t, err)

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

	got, err := threeWayMerge(base, ours, disk)
	require.NoError(t, err)

	require.Equal(t, "b", got.CurrentContext)
}

func TestThreeWayMerge_UntouchedScalarTakesDisk(t *testing.T) {
	base := New()
	base.EnableColor = false
	ours := New()
	ours.EnableColor = false // we didn't touch it
	disk := New()
	disk.EnableColor = true // another session enabled color

	got, err := threeWayMerge(base, ours, disk)
	require.NoError(t, err)

	require.True(t, got.EnableColor, "a field we didn't change must keep the disk value")
}

func TestThreeWayMerge_OurMapValueChangeWins(t *testing.T) {
	base := cfgWithPlatforms("a")
	ours := cfgWithPlatforms("a")
	ours.Platforms["a"] = &Platform{Name: "a", Server: "https://new"} // we edited "a"
	disk := cfgWithPlatforms("a")                                     // disk unchanged

	got, err := threeWayMerge(base, ours, disk)
	require.NoError(t, err)

	require.Equal(t, "https://new", got.Platforms["a"].Server, "our value change to an existing key must win")
}

func TestThreeWayMerge_UntouchedMapValueTakesDisk(t *testing.T) {
	base := cfgWithPlatforms("a")
	ours := cfgWithPlatforms("a") // we didn't touch "a"
	disk := cfgWithPlatforms("a")
	disk.Platforms["a"] = &Platform{Name: "a", Server: "https://disk-changed"} // another session edited "a"

	got, err := threeWayMerge(base, ours, disk)
	require.NoError(t, err)

	require.Equal(t, "https://disk-changed", got.Platforms["a"].Server, "a key we didn't touch must keep disk's in-place change")
}

// helper: a config with one context carrying an environment and an active cluster
func cfgWithContext(name, env, activeKafka string) *Config {
	c := New()
	c.Contexts[name] = &Context{
		Name:                name,
		CurrentEnvironment:  env,
		KafkaClusterContext: &KafkaClusterContext{ActiveKafkaCluster: activeKafka},
	}
	return c
}

func TestThreeWayMerge_MergesConcurrentFieldsOfSameContext(t *testing.T) {
	base := cfgWithContext("ctx", "env-a", "lkc-1")
	ours := cfgWithContext("ctx", "env-b", "lkc-1") // we ran `environment use env-b`
	disk := cfgWithContext("ctx", "env-a", "lkc-2") // another session ran `kafka cluster use lkc-2`

	got, err := threeWayMerge(base, ours, disk)
	require.NoError(t, err)

	require.Equal(t, "env-b", got.Contexts["ctx"].CurrentEnvironment, "our field change must survive")
	require.Equal(t, "lkc-2", got.Contexts["ctx"].KafkaClusterContext.ActiveKafkaCluster, "the concurrent field change on disk must survive")
}

// Arrays are leaves: a changed array takes ours wholesale, an untouched one keeps
// disk's. No element-wise merging, matching the pre-existing atomic map-value merge.
func TestMergeValue_TreatsArraysAtomically(t *testing.T) {
	base := map[string]any{"list": []any{"a", "b"}}
	ours := map[string]any{"list": []any{"a", "b", "c"}} // we appended "c"
	disk := map[string]any{"list": []any{"a", "b"}}      // disk unchanged

	got := mergeValue(base, ours, disk).(map[string]any)

	require.Equal(t, []any{"a", "b", "c"}, got["list"], "a changed array takes ours wholesale")
}

func TestMergeValue_UntouchedArrayTakesDisk(t *testing.T) {
	base := map[string]any{"list": []any{"a"}}
	ours := map[string]any{"list": []any{"a"}}      // unchanged by us
	disk := map[string]any{"list": []any{"a", "b"}} // another writer extended it

	got := mergeValue(base, ours, disk).(map[string]any)

	require.Equal(t, []any{"a", "b"}, got["list"], "an untouched array keeps disk's version")
}

func TestMergeValue_RecursesDeletesAndConcurrentAdds(t *testing.T) {
	base := map[string]any{"keep": "v", "drop": "v"}
	ours := map[string]any{"keep": "v", "mine": "v"}           // we deleted "drop" and added "mine"
	disk := map[string]any{"keep": "v", "drop": "v", "x": "v"} // another writer added "x"

	got := mergeValue(base, ours, disk).(map[string]any)

	require.NotContains(t, got, "drop", "our delete must be applied")
	require.Contains(t, got, "mine", "our add must survive")
	require.Contains(t, got, "x", "a concurrent add must survive")
}
