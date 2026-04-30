package flink

import "github.com/confluentinc/cli/v4/pkg/log"

// GetMapField extracts a typed value of type T from an untyped map.
// cmf-sdk-go represents K8s/FKO-style fields (status, spec, metadata, defaults)
// as map[string]any since their schema is controller-driven and not statically
// known. This is the single safe entry point for reading those fields.
//
// Returns (zero, false) if the key is absent, nil, or not of type T. A type
// mismatch is logged at debug level — that's a server/schema contract
// violation, not a normal absence.
//
// For nested maps, call repeatedly:
//
//	jobStatus, ok := GetMapField[map[string]any](status, "jobStatus", label)
//	if ok {
//		state, _ := GetMapField[string](jobStatus, "state", label)
//	}
//
// contextLabel should identify the resource for actionable debug logs
// (e.g. `compute pool "foo"`).
func GetMapField[T any](m map[string]any, key, contextLabel string) (T, bool) {
	var zero T
	raw, ok := m[key]
	if !ok || raw == nil {
		return zero, false
	}
	v, ok := raw.(T)
	if !ok {
		log.CliLogger.Debugf("%s: %s has unexpected type %T, expected %T", contextLabel, key, raw, zero)
		return zero, false
	}
	return v, true
}
