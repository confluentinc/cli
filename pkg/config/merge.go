package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// threeWayMerge produces the config to persist. Disk ("theirs") is the base onto
// which this process's own changes are overlaid: a field another session changed
// and this process did not touch is preserved. base is the common ancestor (our
// view of the persisted state as of our last load or successful save); ours is
// our current state. Scalars and pointer fields take ours when we changed them
// from base, else disk's. Map fields merge per key AND per nested field, so two
// sessions that changed different fields of the same value (for example a
// context's environment vs. its active Kafka cluster) both survive; a key present
// only on disk (a concurrent add) survives, and a key we deleted (in base, absent
// from ours) is removed.
//
// base and ours must be in the same representation before this runs, or a secret this
// process never touched reads as base != ours (a local change) and overwrites a
// concurrent update. saveLocked aligns them via decryptToMatch. disk's own
// representation does not matter: an untouched field is taken from disk wholesale.
func threeWayMerge(base, ours, disk *Config) (*Config, error) {
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

	var err error
	if out.Platforms, err = mergeMapDeep(base.Platforms, ours.Platforms, disk.Platforms); err != nil {
		return nil, err
	}
	if out.Credentials, err = mergeMapDeep(base.Credentials, ours.Credentials, disk.Credentials); err != nil {
		return nil, err
	}
	if out.Contexts, err = mergeMapDeep(base.Contexts, ours.Contexts, disk.Contexts); err != nil {
		return nil, err
	}
	if out.ContextStates, err = mergeMapDeep(base.ContextStates, ours.ContextStates, disk.ContextStates); err != nil {
		return nil, err
	}
	if out.SavedCredentials, err = mergeMapDeep(base.SavedCredentials, ours.SavedCredentials, disk.SavedCredentials); err != nil {
		return nil, err
	}

	return out, nil
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

// mergeMapDeep three-way-merges one map field, recursing into each value's nested
// structure so concurrent edits to different fields of the same value both
// survive. It works in generic JSON form (map -> tree -> map) so it needs no
// per-type knowledge; base is the common ancestor, ours our current map, disk the
// on-disk map that changes are overlaid onto.
func mergeMapDeep[V any](base, ours, disk map[string]V) (map[string]V, error) {
	baseTree, err := toTree(base)
	if err != nil {
		return nil, err
	}
	oursTree, err := toTree(ours)
	if err != nil {
		return nil, err
	}
	diskTree, err := toTree(disk)
	if err != nil {
		return nil, err
	}

	merged := mergeValue(baseTree, oursTree, diskTree)

	var out map[string]V
	if err := fromTree(merged, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]V{}
	}
	return out, nil
}

// mergeValue three-way-merges one JSON value. base is the common ancestor, ours
// our value, disk the on-disk value. Objects recurse key by key so nested
// concurrent edits both survive; any other value (scalar, array, null) is atomic
// and takes ours when we changed it from base, else disk's.
func mergeValue(base, ours, disk any) any {
	oursObj, oursIsObj := ours.(map[string]any)
	diskObj, diskIsObj := disk.(map[string]any)
	if !oursIsObj || !diskIsObj {
		if !reflect.DeepEqual(base, ours) {
			return ours
		}
		return disk
	}
	baseObj, _ := base.(map[string]any) // nil when base is absent or not an object

	keys := make(map[string]struct{}, len(baseObj)+len(oursObj)+len(diskObj))
	for k := range baseObj {
		keys[k] = struct{}{}
	}
	for k := range oursObj {
		keys[k] = struct{}{}
	}
	for k := range diskObj {
		keys[k] = struct{}{}
	}

	out := make(map[string]any, len(keys))
	for k := range keys {
		bv, inBase := baseObj[k]
		ov, inOurs := oursObj[k]
		dv, inDisk := diskObj[k]

		switch {
		case !inOurs:
			// We do not have the key. If it was in our ancestor we deleted it, and
			// our delete wins; otherwise it is a concurrent add on disk, so keep it.
			if !inBase {
				out[k] = dv
			}
		case inDisk:
			// Present on both sides: recurse so nested concurrent edits both survive.
			out[k] = mergeValue(bv, ov, dv)
		default:
			// Disk lacks it. Keep ours if we added it (not in base) or changed it;
			// if it is unchanged from base, disk deleted it, so respect that.
			if !inBase || !reflect.DeepEqual(bv, ov) {
				out[k] = ov
			}
		}
	}
	return out
}

// toTree renders v as a generic JSON value. UseNumber keeps numbers as exact
// decimal strings so a large integer id is not corrupted by a float64 round-trip
// and so equal numbers compare equal.
func toTree(v any) (any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("unable to convert config to merge tree: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var tree any
	if err := decoder.Decode(&tree); err != nil {
		return nil, fmt.Errorf("unable to convert config to merge tree: %w", err)
	}
	return tree, nil
}

// fromTree unmarshals a generic JSON value produced by mergeValue back into out.
func fromTree(tree, out any) error {
	data, err := json.Marshal(tree)
	if err != nil {
		return fmt.Errorf("unable to convert merge tree to config: %w", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("unable to convert merge tree to config: %w", err)
	}
	return nil
}
