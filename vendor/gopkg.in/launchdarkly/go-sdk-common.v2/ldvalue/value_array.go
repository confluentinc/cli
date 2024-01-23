package ldvalue

import (
	"encoding/json"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream" //nolint:staticcheck // using a deprecated API

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

// we reuse this for all non-nil zero-length ValueArray instances
var emptyArray = []Value{} //nolint:gochecknoglobals

// ValueArray is an immutable array of Value values.
//
// This is used internally to hold the contents of a JSON array in a Value. You can also use it
// separately in any context where you know that the data must be array-like, rather than any of the
// other types that a Value can be.
//
// The wrapped slice is not directly accessible, so it cannot be modified. You can obtain a copy of
// it with AsSlice() if necessary.
//
// Like a Go slice, there is a distinction between an array in a nil state-- which is the zero
// value of ValueArray{}-- and a non-nil aray that is empty. The former is represented in JSON as a
// null; the latter is an empty JSON array [].
type ValueArray struct {
	data []Value
}

// ValueArrayBuilder is a builder created by ValueArrayBuild(), for creating immutable JSON arrays.
//
// A ValueArrayBuilder should not be accessed by multiple goroutines at once.
type ValueArrayBuilder interface {
	// Add appends an element to the array builder.
	Add(value Value) ValueArrayBuilder
	// AddAllFromArray appends all elements from an existing ValueArray.
	AddAllFromValueArray(ValueArray) ValueArrayBuilder
	// Build creates a Value containing the previously added array elements. Continuing to modify the
	// same builder by calling Add after that point does not affect the returned array.
	Build() ValueArray
}

type valueArrayBuilderImpl struct {
	copyOnWrite bool
	output      []Value
}

// ValueArrayBuild creates a builder for constructing an immutable ValueArray.
//
//     ValueArray := ldvalue.ValueArrayBuild().Add(ldvalue.Int(100)).Add(ldvalue.Int(200)).Build()
func ValueArrayBuild() ValueArrayBuilder {
	return &valueArrayBuilderImpl{}
}

// ValueArrayBuildWithCapacity creates a builder for constructing an immutable ValueArray.
//
// The capacity parameter is the same as the capacity of a slice, allowing you to preallocate space
// if you know the number of elements; otherwise you can pass zero.
//
//     arrayValue := ldvalue.ValueArrayBuildWithCapacity(2).Add(ldvalue.Int(100)).Add(ldvalue.Int(200)).Build()
func ValueArrayBuildWithCapacity(capacity int) ValueArrayBuilder {
	return &valueArrayBuilderImpl{output: make([]Value, 0, capacity)}
}

// ValueArrayBuildFromArray creates a builder for constructing an immutable ValueArray, initializing it
// from an existing ValueArray.
//
// The builder has copy-on-write behavior, so if you make no changes before calling Build(), the
// original array is used as-is.
func ValueArrayBuildFromArray(a ValueArray) ValueArrayBuilder {
	return &valueArrayBuilderImpl{output: a.data, copyOnWrite: true}
}

func (b *valueArrayBuilderImpl) Add(value Value) ValueArrayBuilder {
	if b.copyOnWrite {
		n := len(b.output)
		newSlice := make([]Value, n, n+1)
		copy(newSlice[0:n], b.output)
		b.output = newSlice
		b.copyOnWrite = false
	}
	if b.output == nil {
		b.output = make([]Value, 0, 1)
	}
	b.output = append(b.output, value)
	return b
}

func (b *valueArrayBuilderImpl) AddAllFromValueArray(a ValueArray) ValueArrayBuilder {
	for _, v := range a.data {
		b.Add(v)
	}
	return b
}

func (b *valueArrayBuilderImpl) Build() ValueArray {
	if b.output == nil {
		return ValueArray{emptyArray}
	}
	b.copyOnWrite = true
	return ValueArray{b.output}
}

// ValueArrayOf creates a ValueArray from a list of Values.
//
// This requires a slice copy to ensure immutability; otherwise, an existing slice could be passed
// using the spread operator, and then modified. However, since Value is itself immutable, it does
// not need to deep-copy each item.
func ValueArrayOf(items ...Value) ValueArray {
	// ValueArrayOf() with no parameters will pass nil rather than a zero-length slice; logically we
	// still want it to create a non-nil array.
	if items == nil {
		return ValueArray{emptyArray}
	}
	return CopyValueArray(items)
}

// CopyValueArray copies an existing ordinary map to a ValueArray.
//
// If the parameter is nil, an uninitialized ValueArray{} is returned instead of a zero-length array.
func CopyValueArray(data []Value) ValueArray {
	if data == nil {
		return ValueArray{}
	}
	if len(data) == 0 {
		return ValueArray{emptyArray}
	}
	a := make([]Value, len(data))
	copy(a, data)
	return ValueArray{data: a}
}

// CopyArbitraryValueArray copies an existing ordinary map of interface{} values to a ValueArray. The
// behavior for each value is the same as CopyArbitraryValue.
//
// If the parameter is nil, an uninitialized ValueArray{} is returned instead of a zero-length map.
func CopyArbitraryValueArray(data []interface{}) ValueArray {
	if data == nil {
		return ValueArray{}
	}
	a := make([]Value, len(data))
	for i, v := range data {
		a[i] = CopyArbitraryValue(v)
	}
	return ValueArray{data: a}
}

// IsDefined returns true if the array is non-nil.
func (a ValueArray) IsDefined() bool {
	return a.data != nil
}

// Count returns the number of elements in the array. For an uninitialized ValueArray{}, this is zero.
func (a ValueArray) Count() int {
	return len(a.data)
}

// AsValue converts the ValueArray to a Value which is either Null() or an array. This does not
// cause any new allocations.
func (a ValueArray) AsValue() Value {
	if a.data == nil {
		return Null()
	}
	return Value{valueType: ArrayType, arrayValue: a}
}

// Get gets a value from the array by index.
//
// If the index is out of range, it returns Null().
func (a ValueArray) Get(index int) Value {
	if index < 0 || index >= len(a.data) {
		return Null()
	}
	return a.data[index]
}

// TryGet gets a value from the map by index, with a second return value of true if successful.
//
// If the index is out of range, it returns (Null(), false).
func (a ValueArray) TryGet(index int) (Value, bool) {
	if index < 0 || index >= len(a.data) {
		return Null(), false
	}
	return a.data[index], true
}

// AsSlice returns a copy of the wrapped data as a simple Go slice whose values are of type Value.
//
// For an uninitialized ValueArray{}, this returns nil.
func (a ValueArray) AsSlice() []Value {
	if a.data == nil {
		return nil
	}
	ret := make([]Value, len(a.data))
	copy(ret, a.data)
	return ret
}

// AsArbitraryValueSlice returns a copy of the wrapped data as a simple Go slice whose values are
// of type interface{}. The behavior for each value is the same as Value.AsArbitraryValue().
//
// For an uninitialized ValueArray{}, this returns nil.
func (a ValueArray) AsArbitraryValueSlice() []interface{} {
	if a.data == nil {
		return nil
	}
	ret := make([]interface{}, len(a.data))
	for i, v := range a.data {
		ret[i] = v.AsArbitraryValue()
	}
	return ret
}

// Equal returns true if the two arrays are deeply equal. Nil and zero-length arrays are not considered
// equal to each other.
func (a ValueArray) Equal(other ValueArray) bool {
	if len(a.data) != len(other.data) || a.IsDefined() != other.IsDefined() {
		return false
	}
	for i, v := range a.data {
		if !v.Equal(other.data[i]) {
			return false
		}
	}
	return true
}

// Enumerate calls a function for each value within a ValueArray, passing the index with each.
//
// The return value of fn is true to continue enumerating, false to stop.
func (a ValueArray) Enumerate(fn func(index int, value Value) bool) {
	for i, v := range a.data {
		if !fn(i, v) {
			return
		}
	}
}

// Transform applies a transformation function to a ValueArray, returning a new ValueArray.
//
// The behavior is as follows:
//
// If the input value is nil or zero-length, the result is identical and the function is not called.
//
// Otherwise, fn is called for each value. It should return a transformed value and true, or else
// return false for the second return value if the property should be dropped.
func (a ValueArray) Transform(fn func(index int, value Value) (Value, bool)) ValueArray {
	if len(a.data) == 0 {
		return a
	}
	ret := a.data
	startedNewSlice := false
	for i, v := range a.data {
		transformedValue, ok := fn(i, v)
		modified := !ok || !transformedValue.Equal(v)
		if modified && !startedNewSlice {
			// This is the first change we've seen, so we should start building a new slice and
			// retroactively add any values to it that already passed the test without changes.
			startedNewSlice = true
			ret = make([]Value, i, len(a.data))
			copy(ret, a.data)
		}
		if startedNewSlice && ok {
			ret = append(ret, transformedValue)
		}
	}
	return ValueArray{ret}
}

// String converts the value to a string representation, equivalent to JSONString().
//
// This method is provided because it is common to use the Stringer interface as a quick way to
// summarize the contents of a value. The simplest way to do so in this case is to use the JSON
// representation.
func (a ValueArray) String() string {
	return a.JSONString()
}

// JSONString returns the JSON representation of the map.
func (a ValueArray) JSONString() string {
	bytes, _ := a.MarshalJSON()
	// It shouldn't be possible for marshalling to fail, because Value can only contain
	// JSON-compatible types. But if it somehow did fail, bytes will be nil and we'll return
	// an empty tring.
	return string(bytes)
}

// MarshalJSON converts the ValueArray to its JSON representation.
//
// Like a Go slice, a ValueArray in an uninitialized/nil state produces a JSON null rather than an empty [].
func (a ValueArray) MarshalJSON() ([]byte, error) {
	if a.data == nil {
		return nullAsJSONBytes, nil
	}
	return json.Marshal(a.data)
}

// UnmarshalJSON parses a ValueArray from JSON.
func (a *ValueArray) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, a)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Unmarshal.
// See the jsonstream package for more details.
func (a *ValueArray) ReadFromJSONReader(r *jreader.Reader) {
	arr := r.ArrayOrNull()
	a.readFromJSONArray(r, &arr)
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// The JSON output format is identical to what is produced by json.Marshal, but this implementation is
// more efficient when building output with jsonstream. See the jsonstream package for more details.
//
// Like a Go slice, a ValueArray in an uninitialized/nil state produces a JSON null rather than an empty [].
func (a ValueArray) WriteToJSONWriter(w *jwriter.Writer) {
	if a.data == nil {
		w.Null()
		return
	}
	arr := w.Array()
	for _, v := range a.data {
		v.WriteToJSONWriter(w)
	}
	arr.End()
}

// WriteToJSONBuffer provides JSON serialization for use with the deprecated jsonstream API.
//
// Deprecated: this method is provided for backward compatibility. The LaunchDarkly SDK no longer
// uses this API; instead it uses the newer https://github.com/launchdarkly/go-jsonstream.
//
// Like a Go slice, a ValueArray in an uninitialized/nil state produces a JSON null rather than an empty [].
func (a ValueArray) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	if a.data == nil {
		j.WriteNull()
		return
	}
	jsonstream.WriteToJSONBufferThroughWriter(a, j)
}

func (a *ValueArray) readFromJSONArray(r *jreader.Reader, arr *jreader.ArrayState) {
	if r.Error() != nil {
		return
	}
	if !arr.IsDefined() {
		*a = ValueArray{}
		return
	}
	var ab ValueArrayBuilder
	for arr.Next() {
		if ab == nil {
			ab = ValueArrayBuildWithCapacity(2)
		}
		var vv Value
		vv.ReadFromJSONReader(r)
		ab.Add(vv)
	}
	if r.Error() == nil {
		if ab == nil {
			*a = ValueArrayOf()
		} else {
			*a = ab.Build()
		}
	}
}
