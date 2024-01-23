package ldvalue

import (
	"encoding/json"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream" //nolint:staticcheck // using a deprecated API

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

// OptionalString represents a string that may or may not have a value. This is similar to using a
// string pointer to distinguish between an empty string and nil, but it is safer because it does
// not expose a pointer to any mutable value.
//
// Unlike Value, which can contain a value of any JSON-compatible type, OptionalString either
// contains a string or nothing. To create an instance with a string value, use NewOptionalString.
// There is no corresponding method for creating an instance with no value; simply use the empty
// literal OptionalString{}.
//
//     os1 := NewOptionalString("this has a value")
//     os2 := NewOptionalString("") // this has a value which is an empty string
//     os2 := OptionalString{} // this does not have a value
//
// This can also be used as a convenient way to construct a string pointer within an expression.
// For instance, this example causes myStringPointer to point to the string "x":
//
//     var myStringPointer *string = NewOptionalString("x").AsPointer()
type OptionalString struct {
	value    string
	hasValue bool
}

// NewOptionalString constructs an OptionalString that has a string value.
//
// There is no corresponding method for creating an OptionalString with no value; simply use
// the empty literal OptionalString{}.
func NewOptionalString(value string) OptionalString {
	return OptionalString{value: value, hasValue: true}
}

// NewOptionalStringFromPointer constructs an OptionalString from a string pointer. If the pointer
// is non-nil, then the OptionalString copies its value; otherwise the OptionalString has no value.
func NewOptionalStringFromPointer(valuePointer *string) OptionalString {
	if valuePointer == nil {
		return OptionalString{hasValue: false}
	}
	return OptionalString{value: *valuePointer, hasValue: true}
}

// IsDefined returns true if the OptionalString contains a string value, or false if it has no value.
func (o OptionalString) IsDefined() bool {
	return o.hasValue
}

// StringValue returns the OptionalString's value, or an empty string if it has no value.
func (o OptionalString) StringValue() string {
	return o.value
}

// Get is a combination of StringValue and IsDefined. If the OptionalString contains a string value,
// it returns that value and true; otherwise it returns an empty string and false.
func (o OptionalString) Get() (string, bool) {
	return o.value, o.hasValue
}

// OrElse returns the OptionalString's value if it has one, or else the specified fallback value.
func (o OptionalString) OrElse(valueIfEmpty string) string {
	if o.hasValue {
		return o.value
	}
	return valueIfEmpty
}

// OnlyIfNonEmptyString returns the same OptionalString unless it contains an empty string (""), in
// which case it returns an OptionalString that has no value.
func (o OptionalString) OnlyIfNonEmptyString() OptionalString {
	if o.hasValue && o.value == "" {
		return OptionalString{}
	}
	return o
}

// AsPointer returns the OptionalString's value as a string pointer if it has a value, or
// nil otherwise.
//
// The string value, if any, is copied rather than returning to a pointer to the internal field.
func (o OptionalString) AsPointer() *string {
	if o.hasValue {
		s := o.value
		return &s
	}
	return nil
}

// AsValue converts the OptionalString to a Value, which is either Null() or a string value.
func (o OptionalString) AsValue() Value {
	if o.hasValue {
		return String(o.value)
	}
	return Null()
}

// String is a debugging convenience method that returns a description of the OptionalString.
// This is either the same as its string value, "[empty]" if it has a string value that is empty,
// or "[none]" if it has no value.
//
// Since String() is used by methods like fmt.Printf, if you want to use the actual string value
// of the OptionalString instead of this special representation, you should call StringValue():
//
//     s := OptionalString{}
//     fmt.Printf("it is '%s'", s) // prints "it is '[none]'"
//     fmt.Printf("it is '%s'", s.StringValue()) // prints "it is ''"
func (o OptionalString) String() string {
	if o.hasValue {
		if o.value == "" {
			return "[empty]"
		}
		return o.value
	}
	return noneDescription
}

// MarshalJSON converts the OptionalString to its JSON representation.
//
// The output will be either a JSON string or null. Note that the "omitempty" tag for a struct
// field will not cause an empty OptionalString field to be omitted; it will be output as null.
// If you want to completely omit a JSON property when there is no value, it must be a string
// pointer instead of an OptionalString; use the AsPointer() method to get a pointer.
func (o OptionalString) MarshalJSON() ([]byte, error) {
	if o.hasValue {
		return json.Marshal(o.value)
	}
	return nullAsJSONBytes, nil
}

// UnmarshalJSON parses an OptionalString from JSON.
//
// The input must be either a JSON string or null.
func (o *OptionalString) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, o)
}

// MarshalText implements the encoding.TextMarshaler interface.
//
// Marshaling an empty OptionalString{} will return nil, rather than a zero-length slice.
func (o OptionalString) MarshalText() ([]byte, error) {
	if o.hasValue {
		return []byte(o.value), nil
	}
	return nil, nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// This allows OptionalString to be used with packages that can parse text content, such as gcfg.
//
// If the byte slice is nil, rather than zero-length, it will set the target value to an empty
// OptionalString{}. Otherwise, it will set it to a string value.
func (o *OptionalString) UnmarshalText(text []byte) error {
	if text == nil {
		*o = OptionalString{}
	} else {
		*o = NewOptionalString(string(text))
	}
	return nil
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Unmarshal.
// See https://github.com/launchdarkly/go-jsonstream for more details.
func (o *OptionalString) ReadFromJSONReader(r *jreader.Reader) {
	val, nonNull := r.StringOrNull()
	if r.Error() == nil {
		if nonNull {
			*o = NewOptionalString(val)
		} else {
			*o = OptionalString{}
		}
	}
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Marshal.
// See https://github.com/launchdarkly/go-jsonstream for more details.
func (o OptionalString) WriteToJSONWriter(w *jwriter.Writer) {
	if o.hasValue {
		w.String(o.value)
	} else {
		w.Null()
	}
}

// WriteToJSONBuffer provides JSON serialization for use with the deprecated jsonstream API.
//
// Deprecated: this method is provided for backward compatibility. The LaunchDarkly SDK no longer
// uses this API; instead it uses the newer https://github.com/launchdarkly/go-jsonstream.
func (o OptionalString) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	o.AsValue().WriteToJSONBuffer(j)
}
