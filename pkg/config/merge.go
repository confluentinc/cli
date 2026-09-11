package config

import "reflect"

// threeWayMerge produces the config to persist. Disk ("theirs") is the base:
// fields another session changed and this process did not touch are preserved.
// This process's own changes, computed as base ("common ancestor") vs ours,
// are then overlaid. For scalars, a value we changed from base wins; a value we
// left alone keeps disk's. For maps, our adds/updates win and our deletes (in
// base, absent from ours) are removed from disk, while keys only present on disk
// survive.
//
// Field-class table (every persisted field of Config must appear here):
//
//	scalars:  DisableFeatureFlags, DisablePlugins, DisablePluginsOnceWindows,
//	          DisableUpdateCheck, EnableColor, CurrentContext, DisablePluginsOnce
//	*time:    LastUpdateCheckAt
//	*struct:  LocalPorts
//	maps:     Platforms, Credentials, Contexts, ContextStates, SavedCredentials
func threeWayMerge(base, ours, disk *Config) *Config {
	out := disk

	out.DisableFeatureFlags = mergeScalar(base.DisableFeatureFlags, ours.DisableFeatureFlags, disk.DisableFeatureFlags)
	out.DisablePlugins = mergeScalar(base.DisablePlugins, ours.DisablePlugins, disk.DisablePlugins)
	out.DisablePluginsOnceWindows = mergeScalar(base.DisablePluginsOnceWindows, ours.DisablePluginsOnceWindows, disk.DisablePluginsOnceWindows)
	out.DisableUpdateCheck = mergeScalar(base.DisableUpdateCheck, ours.DisableUpdateCheck, disk.DisableUpdateCheck)
	out.EnableColor = mergeScalar(base.EnableColor, ours.EnableColor, disk.EnableColor)
	out.CurrentContext = mergeScalar(base.CurrentContext, ours.CurrentContext, disk.CurrentContext)
	out.DisablePluginsOnce = mergeScalar(base.DisablePluginsOnce, ours.DisablePluginsOnce, disk.DisablePluginsOnce)

	out.LastUpdateCheckAt = mergePtr(base.LastUpdateCheckAt, ours.LastUpdateCheckAt, disk.LastUpdateCheckAt)
	out.LocalPorts = mergePtr(base.LocalPorts, ours.LocalPorts, disk.LocalPorts)

	out.Platforms = mergeMap(out.Platforms, base.Platforms, ours.Platforms)
	out.Credentials = mergeMap(out.Credentials, base.Credentials, ours.Credentials)
	out.Contexts = mergeMap(out.Contexts, base.Contexts, ours.Contexts)
	out.ContextStates = mergeMap(out.ContextStates, base.ContextStates, ours.ContextStates)
	out.SavedCredentials = mergeMap(out.SavedCredentials, base.SavedCredentials, ours.SavedCredentials)

	return out
}

// mergeScalar returns ours if this process changed it from base, else disk's.
func mergeScalar[T comparable](base, ours, disk T) T {
	if ours != base {
		return ours
	}
	return disk
}

// mergePtr treats a pointer field like a scalar, comparing pointed-to values.
func mergePtr[T any](base, ours, disk *T) *T {
	if !reflect.DeepEqual(base, ours) {
		return ours
	}
	return disk
}

// mergeMap overlays this process's adds/updates/deletes onto the disk map.
// Keys only on disk (a concurrent add) survive; keys we deleted (in base, not in
// ours) are removed; keys we added or changed take our value.
func mergeMap[V any](out, base, ours map[string]V) map[string]V {
	if out == nil {
		out = map[string]V{}
	}
	for k := range base {
		if _, keptByUs := ours[k]; !keptByUs {
			delete(out, k)
		}
	}
	for k, v := range ours {
		bv, inBase := base[k]
		if !inBase || !reflect.DeepEqual(bv, v) {
			out[k] = v
		}
	}
	return out
}
