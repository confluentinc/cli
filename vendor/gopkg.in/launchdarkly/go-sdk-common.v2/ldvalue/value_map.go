package ldvalue

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream" //nolint:staticcheck // using a deprecated API

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

// we reuse this for all non-nil zero-length ValueMap instances
var emptyMap = map[string]Value{} //nolint:gochecknoglobals

// ValueMap is an immutable map of string keys to Value values.
//
// This is used internally to hold the properties of a JSON object in a Value. You can also use it
// separately in any context where you know that the data must be map-like, rather than any of the
// other types that a Value can be.
//
// The wrapped map is not directly accessible, so it cannot be modified. You can obtain a copy of
// it with AsMap() if necessary.
//
// Like a Go map, there is a distinction between a map in a nil state-- which is the zero value of
// ValueMap{}-- and a non-nil map that is empty. The former is represented in JSON as a null; the
// latter is an empty JSON object {}.
type ValueMap struct {
	data map[string]Value
}

// ValueMapBuilder is a builder created by ValueMapBuild(), for creating immutable JSON objects.
//
// A ValueMapBuilder should not be accessed by multiple goroutines at once.
type ValueMapBuilder interface {
	// Set sets a key-value pair in the map builder.
	Set(key string, value Value) ValueMapBuilder
	// SetAllFromValueMap copies all key-value pairs from an existing ValueMap.
	SetAllFromValueMap(ValueMap) ValueMapBuilder
	// Remove unsets a key if it was set.
	Remove(key string) ValueMapBuilder
	// Build creates a ValueMap containing the previously specified key-value pairs. Continuing to
	// modify the same builder by calling Set after that point does not affect the returned ValueMap.
	Build() ValueMap
}

type valueMapBuilderImpl struct {
	copyOnWrite bool
	output      map[string]Value
}

// ValueMapBuild creates a builder for constructing an immutable ValueMap.
//
//     valueMap := ldvalue.ValueMapBuild().Set("a", ldvalue.Int(100)).Set("b", ldvalue.Int(200)).Build()
func ValueMapBuild() ValueMapBuilder {
	return &valueMapBuilderImpl{}
}

// ValueMapBuildWithCapacity creates a builder for constructing an immutable ValueMap.
//
// The capacity parameter is the same as the capacity of a map, allowing you to preallocate space
// if you know the number of elements; otherwise you can pass zero.
//
//     objValue := ldvalue.ObjectBuildWithCapacity(2).Set("a", ldvalue.Int(100)).Set("b", ldvalue.Int(200)).Build()
func ValueMapBuildWithCapacity(capacity int) ValueMapBuilder {
	return &valueMapBuilderImpl{output: make(map[string]Value, capacity)}
}

// ValueMapBuildFromMap creates a builder for constructing an immutable ValueMap, initializing it
// from an existing ValueMap.
//
// The builder has copy-on-write behavior, so if you make no changes before calling Build(), the
// original map is used as-is.
func ValueMapBuildFromMap(m ValueMap) ValueMapBuilder {
	return &valueMapBuilderImpl{output: m.data, copyOnWrite: true}
}

func (b *valueMapBuilderImpl) Set(name string, value Value) ValueMapBuilder {
	if b.copyOnWrite {
		b.output = copyValueMap(b.output)
		b.copyOnWrite = false
	}
	if b.output == nil {
		b.output = make(map[string]Value, 1)
	}
	b.output[name] = value
	return b
}

func (b *valueMapBuilderImpl) Remove(name string) ValueMapBuilder {
	if b.copyOnWrite {
		b.output = copyValueMap(b.output)
		b.copyOnWrite = false
	}
	if b.output != nil {
		delete(b.output, name)
	}
	return b
}

func copyValueMap(m map[string]Value) map[string]Value {
	ret := make(map[string]Value, len(m))
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

func (b *valueMapBuilderImpl) SetAllFromValueMap(m ValueMap) ValueMapBuilder {
	for k, v := range m.data {
		b.Set(k, v)
	}
	return b
}

func (b *valueMapBuilderImpl) Build() ValueMap {
	if b.output == nil {
		return ValueMap{emptyMap}
	}
	b.copyOnWrite = true
	return ValueMap{b.output}
}

// CopyValueMap copies an existing ordinary map to a ValueMap.
//
// If the parameter is nil, an uninitialized ValueMap{} is returned instead of a zero-length map.
func CopyValueMap(data map[string]Value) ValueMap {
	if data == nil {
		return ValueMap{}
	}
	if len(data) == 0 {
		return ValueMap{emptyMap}
	}
	m := make(map[string]Value, len(data))
	for k, v := range data {
		m[k] = v
	}
	return ValueMap{data: m}
}

// CopyArbitraryValueMap copies an existing ordinary map of interface{} values to a ValueMap. The
// behavior for each value is the same as CopyArbitraryValue.
//
// If the parameter is nil, an uninitialized ValueMap{} is returned instead of a zero-length map.
func CopyArbitraryValueMap(data map[string]interface{}) ValueMap {
	if data == nil {
		return ValueMap{}
	}
	m := make(map[string]Value, len(data))
	for k, v := range data {
		m[k] = CopyArbitraryValue(v)
	}
	return ValueMap{data: m}
}

// IsDefined returns true if the map is non-nil.
func (m ValueMap) IsDefined() bool {
	return m.data != nil
}

// Count returns the number of keys in the map. For an uninitialized ValueMap{}, this is zero.
func (m ValueMap) Count() int {
	return len(m.data)
}

// AsValue converts the ValueMap to a Value which is either Null() or an object. This does not
// cause any new allocations.
func (m ValueMap) AsValue() Value {
	if m.data == nil {
		return Null()
	}
	return Value{valueType: ObjectType, objectValue: m}
}

// Get gets a value from the map by key.
//
// If the key is not found, it returns Null().
func (m ValueMap) Get(key string) Value {
	return m.data[key]
}

// TryGet gets a value from the map by key, with a second return value of true if successful.
//
// If the key is not found, it returns (Null(), false).
func (m ValueMap) TryGet(key string) (Value, bool) {
	ret, ok := m.data[key]
	return ret, ok
}

// Keys returns the keys of a the map as a slice.
//
// The method copies the keys. For an uninitialized ValueMap{}, it returns nil.
func (m ValueMap) Keys() []string {
	if m.data == nil {
		return nil
	}
	ret := make([]string, len(m.data))
	i := 0
	for key := range m.data {
		ret[i] = key
		i++
	}
	return ret
}

// AsMap returns a copy of the wrapped data as a simple Go map whose values are of type Value.
//
// For an uninitialized ValueMap{}, this returns nil.
func (m ValueMap) AsMap() map[string]Value {
	if m.data == nil {
		return nil
	}
	ret := make(map[string]Value, len(m.data))
	for k, v := range m.data {
		ret[k] = v
	}
	return ret
}

// AsArbitraryValueMap returns a copy of the wrapped data as a simple Go map whose values are of type
// interface{}. The behavior for each value is the same as Value.AsArbitraryValue().
//
// For an uninitialized ValueMap{}, this returns nil.
func (m ValueMap) AsArbitraryValueMap() map[string]interface{} {
	if m.data == nil {
		return nil
	}
	ret := make(map[string]interface{}, len(m.data))
	for k, v := range m.data {
		ret[k] = v.AsArbitraryValue()
	}
	return ret
}

// Equal returns true if the two maps are deeply equal. Nil and zero-length maps are not considered
// equal to each other.
func (m ValueMap) Equal(other ValueMap) bool {
	if len(m.data) != len(other.data) || m.IsDefined() != other.IsDefined() {
		return false
	}
	for k, v := range m.data {
		v1, ok := other.data[k]
		if !ok || !v.Equal(v1) {
			return false
		}
	}
	return true
}

// Enumerate calls a function for each key-value pair within a ValueMap. The ordering is undefined
// since map iteration in Go is non-deterministic.
//
// The return value of fn is true to continue enumerating, false to stop.
func (m ValueMap) Enumerate(fn func(key string, value Value) bool) {
	for k, v := range m.data {
		if !fn(k, v) {
			return
		}
	}
}

// Transform applies a transformation function to a ValueMap, returning a new ValueMap.
//
// The behavior is as follows:
//
// If the input value is nil or zero-length, the result is identical and the function is not called.
//
// Otherwise, fn is called for each key-value pair. It should return a transformed key-value pair
// and true, or else return false for the third return value if the property should be dropped.
func (m ValueMap) Transform(fn func(key string, value Value) (string, Value, bool)) ValueMap {
	if len(m.data) == 0 {
		return m
	}
	ret := m.data
	startedNewMap := false
	seenKeys := make([]string, 0, len(m.data))
	for k, v := range m.data {
		resultKey, resultValue, ok := fn(k, v)
		modified := !ok || resultKey != k || !resultValue.Equal(v)
		if modified && !startedNewMap {
			// This is the first change we've seen, so we should start building a new map and
			// retroactively add any values to it that already passed the test without changes.
			startedNewMap = true
			ret = make(map[string]Value, len(m.data))
			for _, seenKey := range seenKeys {
				ret[seenKey] = m.data[seenKey]
			}
		} else {
			seenKeys = append(seenKeys, k)
		}
		if startedNewMap && ok {
			ret[k] = resultValue
		}
	}
	return ValueMap{ret}
}

// String converts the value to a map representation, equivalent to JSONString().
//
// This method is provided because it is common to use the Stringer interface as a quick way to
// summarize the contents of a value. The simplest way to do so in this case is to use the JSON
// representation.
func (m ValueMap) String() string {
	return m.JSONString()
}

// JSONString returns the JSON representation of the map.
func (m ValueMap) JSONString() string {
	bytes, _ := m.MarshalJSON()
	// It shouldn't be possible for marshalling to fail, because Value can only contain
	// JSON-compatible types. But if it somehow did fail, bytes will be nil and we'll return
	// an empty tring.
	return string(bytes)
}

// MarshalJSON converts the ValueMap to its JSON representation.
//
// Like a Go map, a ValueMap in an uninitialized/nil state produces a JSON null rather than an empty {}.
func (m ValueMap) MarshalJSON() ([]byte, error) {
	return jwriter.MarshalJSONWithWriter(m)
}

// UnmarshalJSON parses a ValueMap from JSON.
func (m *ValueMap) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, m)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Unmarshal.
// See the jsonstream package for more details.
func (m *ValueMap) ReadFromJSONReader(r *jreader.Reader) {
	obj := r.ObjectOrNull()
	m.readFromJSONObject(r, &obj)
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// The JSON output format is identical to what is produced by json.Marshal, but this implementation is
// more efficient when building output with JSONBuffer. See the jsonstream package for more details.
//
// Like a Go map, a ValueMap in an uninitialized/nil state produces a JSON null rather than an empty {}.
func (m ValueMap) WriteToJSONWriter(w *jwriter.Writer) {
	if m.data == nil {
		w.Null()
		return
	}
	obj := w.Object()
	for k, vv := range m.data {
		vv.WriteToJSONWriter(obj.Name(k))
	}
	obj.End()
}

// WriteToJSONBuffer provides JSON serialization for use with the deprecated jsonstream API.
//
// Deprecated: this method is provided for backward compatibility. The LaunchDarkly SDK no longer
// uses this API; instead it uses the newer https://github.com/launchdarkly/go-jsonstream.
//
// Like a Go map, a ValueMap in an uninitialized/nil state produces a JSON null rather than an empty {}.
func (m ValueMap) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	jsonstream.WriteToJSONBufferThroughWriter(m, j)
}

func (m *ValueMap) readFromJSONObject(r *jreader.Reader, obj *jreader.ObjectState) {
	if r.Error() != nil {
		return
	}
	if !obj.IsDefined() {
		*m = ValueMap{}
		return
	}
	var mb ValueMapBuilder
	for obj.Next() {
		if mb == nil {
			mb = ValueMapBuild()
		}
		name := obj.Name()
		var vv Value
		vv.ReadFromJSONReader(r)
		mb.Set(string(name), vv)
	}
	if r.Error() == nil {
		if mb == nil {
			*m = ValueMap{data: emptyMap}
		} else {
			*m = mb.Build()
		}
	}
}
