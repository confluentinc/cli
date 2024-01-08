package ldvalue

import (
	"encoding/json"
	"errors"
	"strconv"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream" //nolint:staticcheck // using a deprecated API

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

// This file contains methods for converting Value to and from JSON.

// Parse returns a Value parsed from a JSON string, or Null if it cannot be parsed.
//
// This is simply a shortcut for calling json.Unmarshal and disregarding errors. It is meant for
// use in test scenarios where malformed data is not a concern.
func Parse(jsonData []byte) Value {
	var v Value
	if err := v.UnmarshalJSON(jsonData); err != nil {
		return Null()
	}
	return v
}

// JSONString returns the JSON representation of the value.
func (v Value) JSONString() string {
	// The following is somewhat redundant with json.Marshal, but it avoids the overhead of
	// converting between byte arrays and strings.
	switch v.valueType {
	case NullType:
		return nullAsJSON
	case BoolType:
		if v.boolValue {
			return trueString
		}
		return falseString
	case NumberType:
		if v.IsInt() {
			return strconv.Itoa(int(v.numberValue))
		}
		return strconv.FormatFloat(v.numberValue, 'f', -1, 64)
	}
	// For all other types, we rely on our custom marshaller.
	bytes, _ := json.Marshal(v)
	// It shouldn't be possible for marshalling to fail, because Value can only contain
	// JSON-compatible types. But if it somehow did fail, bytes will be nil and we'll return
	// an empty string.
	return string(bytes)
}

// MarshalJSON converts the Value to its JSON representation.
//
// Note that the "omitempty" tag for a struct field will not cause an empty Value field to be
// omitted; it will be output as null. If you want to completely omit a JSON property when there
// is no value, it must be a pointer; use AsPointer().
func (v Value) MarshalJSON() ([]byte, error) {
	switch v.valueType {
	case NullType:
		return nullAsJSONBytes, nil
	case BoolType:
		if v.boolValue {
			return trueBytes, nil
		}
		return falseBytes, nil
	case NumberType:
		if v.IsInt() {
			return []byte(strconv.Itoa(int(v.numberValue))), nil
		}
		return []byte(strconv.FormatFloat(v.numberValue, 'f', -1, 64)), nil
	case StringType:
		return json.Marshal(v.stringValue)
	case ArrayType:
		return v.arrayValue.MarshalJSON()
	case ObjectType:
		return v.objectValue.MarshalJSON()
	case RawType:
		if len(v.stringValue) == 0 {
			return nullAsJSONBytes, nil
			// we don't check for other kinds of malformed JSON here, but if it was nil/"" we can assume they meant null
		}
		return []byte(v.stringValue), nil
	}
	return nil, errors.New("unknown data type") // should not be possible
}

// UnmarshalJSON parses a Value from JSON.
func (v *Value) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, v)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Unmarshal.
// See the jsonstream package for more details.
func (v *Value) ReadFromJSONReader(r *jreader.Reader) {
	a := r.Any()
	if r.Error() != nil {
		return
	}
	switch a.Kind {
	case jreader.BoolValue:
		*v = Bool(a.Bool)
	case jreader.NumberValue:
		*v = Float64(a.Number)
	case jreader.StringValue:
		*v = String(a.String)
	case jreader.ArrayValue:
		var va ValueArray
		if va.readFromJSONArray(r, &a.Array); r.Error() == nil {
			*v = Value{valueType: ArrayType, arrayValue: va}
		}
	case jreader.ObjectValue:
		var vm ValueMap
		if vm.readFromJSONObject(r, &a.Object); r.Error() == nil {
			*v = Value{valueType: ObjectType, objectValue: vm}
		}
	default:
		*v = Null()
	}
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Marshal.
// See https://github.com/launchdarkly/go-jsonstream for more details.
func (v Value) WriteToJSONWriter(w *jwriter.Writer) {
	switch v.valueType {
	case NullType:
		w.Null()
	case BoolType:
		w.Bool(v.boolValue)
	case NumberType:
		w.Float64(v.numberValue)
	case StringType:
		w.String(v.stringValue)
	case ArrayType:
		v.arrayValue.WriteToJSONWriter(w)
	case ObjectType:
		v.objectValue.WriteToJSONWriter(w)
	case RawType:
		if len(v.stringValue) == 0 {
			w.Null() // we don't check for other kinds of malformed JSON here, but if it was nil/"" we can assume they meant null
		} else {
			w.Raw([]byte(v.stringValue))
		}
	}
}

// WriteToJSONBuffer provides JSON serialization for use with the deprecated jsonstream API.
//
// Deprecated: this method is provided for backward compatibility. The LaunchDarkly SDK no longer
// uses this API; instead it uses the newer https://github.com/launchdarkly/go-jsonstream.
func (v Value) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	jsonstream.WriteToJSONBufferThroughWriter(v, j)
}
