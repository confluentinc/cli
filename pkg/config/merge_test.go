package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/secret"
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

// Two sessions lazily create the same env context from a base that lacks it. Our copy's
// untouched fields are zero values, not edits, so they must not wipe disk's real value.
func TestMergeValue_BothAddedZeroValueDoesNotOverrideDisk(t *testing.T) {
	base := map[string]any{"kafka_environment_contexts": map[string]any{}}
	ours := map[string]any{"kafka_environment_contexts": map[string]any{
		"env-596": map[string]any{"active_kafka": "", "active_kafka_endpoint": ""}, // we never set a cluster
	}}
	disk := map[string]any{"kafka_environment_contexts": map[string]any{
		"env-596": map[string]any{"active_kafka": "lkc-123", "active_kafka_endpoint": ""}, // another session ran `kafka cluster use`
	}}

	got := mergeValue(base, ours, disk).(map[string]any)

	envContext := got["kafka_environment_contexts"].(map[string]any)["env-596"].(map[string]any)
	require.Equal(t, "lkc-123", envContext["active_kafka"], "our zero value must not override disk's concurrent selection")
	require.Equal(t, "", envContext["active_kafka_endpoint"], "a field zero on both sides stays zero")
}

func TestMergeValue_BothAddedNonZeroValuesKeepOurs(t *testing.T) {
	base := map[string]any{}
	ours := map[string]any{"env": map[string]any{"active_kafka": "lkc-ours"}}
	disk := map[string]any{"env": map[string]any{"active_kafka": "lkc-disk"}}

	got := mergeValue(base, ours, disk).(map[string]any)

	require.Equal(t, "lkc-ours", got["env"].(map[string]any)["active_kafka"], "a real value we set still wins a concurrent add")
}

// The zero rule applies only when the ancestor lacks the key: clearing a value base held
// (logout, cluster delete) is a real edit and must still win.
func TestMergeValue_ClearingAValueBaseHeldStillWins(t *testing.T) {
	for _, diskValue := range []string{"lkc-1", "lkc-2"} {
		t.Run(diskValue, func(t *testing.T) {
			base := map[string]any{"env": map[string]any{"active_kafka": "lkc-1"}}
			ours := map[string]any{"env": map[string]any{"active_kafka": ""}} // we cleared it
			disk := map[string]any{"env": map[string]any{"active_kafka": diskValue}}

			got := mergeValue(base, ours, disk).(map[string]any)

			require.Equal(t, "", got["env"].(map[string]any)["active_kafka"], "our clear of a value base held must win")
		})
	}
}

func TestIsZeroJSON(t *testing.T) {
	tests := []struct {
		value any
		want  bool
	}{
		{nil, true},
		{"", true},
		{false, true},
		{json.Number("0"), true},
		{json.Number("0.0"), true},
		{[]any{}, true},
		{map[string]any{}, true},
		{"x", false},
		{true, false},
		{json.Number("1"), false},
		{[]any{"x"}, false},
		{map[string]any{"k": "v"}, false},
	}
	for _, test := range tests {
		require.Equal(t, test.want, isZeroJSON(test.value), "isZeroJSON(%#v)", test.value)
	}
}

// A state both sides created is one login's crypto unit: both tokens share its salt and
// nonce. A field-wise merge that keeps our salt but takes disk's refresh ciphertext (our
// refresh token is empty) produces a state that can never be decrypted.
func TestThreeWayMerge_BothAddedContextStateStaysDecryptable(t *testing.T) {
	salt, nonce, err := secret.GenerateSaltAndNonce()
	require.NoError(t, err)
	base := New()
	ours := New()
	ours.Contexts["ctx"] = &Context{Name: "ctx"}
	ours.ContextStates["ctx"] = &ContextState{AuthToken: "header.payload.ours", Salt: salt, Nonce: nonce}
	disk := New()
	disk.Contexts["ctx"] = &Context{Name: "ctx"}
	disk.ContextStates["ctx"] = encryptedState(t, "header.payload.signature", "v1.refreshtoken")

	got, err := threeWayMerge(base, ours, disk)
	require.NoError(t, err)

	require.NoError(t, got.ContextStates["ctx"].DecryptAuthRefreshToken("ctx"),
		"a concurrently created state must keep its tokens with the salt and nonce they were encrypted under")
}
