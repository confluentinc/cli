package flink

import "github.com/confluentinc/cli/v4/pkg/log"

// GetMapField extracts a value of type T from an untyped map. cmf-sdk-go
// surfaces controller-driven fields (status, spec, etc.) as map[string]any.
//
// Returns (zero, false) if the key is absent, nil, or not of type T. A type
// mismatch is logged at debug level (contract violation, not a normal absence).
// contextLabel identifies the resource in those logs, e.g. `compute pool "foo"`.
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
