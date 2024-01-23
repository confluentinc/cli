package ldvalue

import (
	"encoding/json"
	"strconv"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream" //nolint:staticcheck // using a deprecated API

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

// OptionalInt represents an int that may or may not have a value. This is similar to using an
// int pointer to distinguish between a zero value and nil, but it is safer because it does not
// expose a pointer to any mutable value.
//
// To create an instance with an int value, use NewOptionalInt. There is no corresponding method
// for creating an instance with no value; simply use the empty literal OptionalInt{}.
//
//     oi1 := NewOptionalInt(1)
//     oi2 := NewOptionalInt(0) // this has a value which is zero
//     oi3 := OptionalInt{}     // this does not have a value
//
// This can also be used as a convenient way to construct an int pointer within an expression.
// For instance, this example causes myIntPointer to point to the int value 2:
//
//     var myIntPointer *int = NewOptionalInt("x").AsPointer()
//
// This type is used in ldreason.EvaluationDetail.VariationIndex, and for other similar fields
// in the LaunchDarkly Go SDK where an int value may or may not be defined.
type OptionalInt struct {
	value    int
	hasValue bool
}

// NewOptionalInt constructs an OptionalInt that has an int value.
//
// There is no corresponding method for creating an OptionalInt with no value; simply use the
// empty literal OptionalInt{}.
func NewOptionalInt(value int) OptionalInt {
	return OptionalInt{value: value, hasValue: true}
}

// NewOptionalIntFromPointer constructs an OptionalInt from an int pointer. If the pointer is
// non-nil, then the OptionalInt copies its value; otherwise the OptionalInt has no value.
func NewOptionalIntFromPointer(valuePointer *int) OptionalInt {
	if valuePointer == nil {
		return OptionalInt{hasValue: false}
	}
	return OptionalInt{value: *valuePointer, hasValue: true}
}

// IsDefined returns true if the OptionalInt contains an int value, or false if it has no value.
func (o OptionalInt) IsDefined() bool {
	return o.hasValue
}

// IntValue returns the OptionalInt's value, or zero if it has no value.
func (o OptionalInt) IntValue() int {
	return o.value
}

// Get is a combination of IntValue and IsDefined. If the OptionalInt contains an int value, it
// returns that value and true; otherwise it returns zero and false.
func (o OptionalInt) Get() (int, bool) {
	return o.value, o.hasValue
}

// OrElse returns the OptionalInt's value if it has one, or else the specified fallback value.
func (o OptionalInt) OrElse(valueIfEmpty int) int {
	if o.hasValue {
		return o.value
	}
	return valueIfEmpty
}

// AsPointer returns the OptionalInt's value as an int pointer if it has a value, or nil
// otherwise.
//
// The int value, if any, is copied rather than returning to a pointer to the internal field.
func (o OptionalInt) AsPointer() *int {
	if o.hasValue {
		v := o.value
		return &v
	}
	return nil
}

// AsValue converts the OptionalInt to a Value, which is either Null() or a number value.
func (o OptionalInt) AsValue() Value {
	if o.hasValue {
		return Int(o.value)
	}
	return Null()
}

// String is a debugging convenience method that returns a description of the OptionalInt. This
// is either a numeric string, or "[none]" if it has no value.
func (o OptionalInt) String() string {
	if o.hasValue {
		return strconv.Itoa(o.value)
	}
	return noneDescription
}

// MarshalJSON converts the OptionalInt to its JSON representation.
//
// The output will be either a JSON number or null. Note that the "omitempty" tag for a struct
// field will not cause an empty OptionalInt field to be omitted; it will be output as null.
// If you want to completely omit a JSON property when there is no value, it must be an int
// pointer instead of an OptionalInt; use the AsPointer() method to get a pointer.
func (o OptionalInt) MarshalJSON() ([]byte, error) {
	if o.hasValue {
		return json.Marshal(o.value)
	}
	return nullAsJSONBytes, nil
}

// UnmarshalJSON parses an OptionalInt from JSON.
//
// The input must be either a JSON number that is an integer or null.
func (o *OptionalInt) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, o)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Unmarshal.
// See https://github.com/launchdarkly/go-jsonstream for more details.
func (o *OptionalInt) ReadFromJSONReader(r *jreader.Reader) {
	val, nonNull := r.IntOrNull()
	if r.Error() == nil {
		if nonNull {
			*o = NewOptionalInt(val)
		} else {
			*o = OptionalInt{}
		}
	}
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Marshal.
// See https://github.com/launchdarkly/go-jsonstream for more details.
func (o OptionalInt) WriteToJSONWriter(w *jwriter.Writer) {
	if o.hasValue {
		w.Int(o.value)
	} else {
		w.Null()
	}
}

// WriteToJSONBuffer provides JSON serialization for use with the deprecated jsonstream API.
//
// Deprecated: this method is provided for backward compatibility. The LaunchDarkly SDK no longer
// uses this API; instead it uses the newer https://github.com/launchdarkly/go-jsonstream.
func (o OptionalInt) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	o.AsValue().WriteToJSONBuffer(j)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (o OptionalInt) MarshalText() ([]byte, error) {
	if o.hasValue {
		return []byte(strconv.Itoa(o.value)), nil
	}
	return []byte(""), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This allows OptionalInt to be used with packages that can parse text content, such as gcfg.
func (o *OptionalInt) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*o = OptionalInt{}
		return nil
	}
	n, err := strconv.Atoi(string(text))
	if err != nil {
		return err
	}
	*o = NewOptionalInt(n)
	return nil
}
