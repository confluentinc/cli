package ldvalue

import (
	"encoding/json"
)

// Value represents any of the data types supported by JSON, all of which can be used for a LaunchDarkly
// feature flag variation or a custom user attribute.
//
// You cannot compare Value instances with the == operator, because the struct may contain a slice.
// Value has an Equal method for this purpose; reflect.DeepEqual will also work.
//
// Value instances are immutable when used by code outside of this package.
type Value struct {
	valueType ValueType
	// Used when the value is a boolean.
	boolValue bool
	// Used when the value is a number.
	numberValue float64
	// Used when the value is a string.
	stringValue string
	// Used when the value is an array, if it was not created from an interface{}.
	arrayValue ValueArray
	// Used when the value is an object, if it was not created from an interface{}.
	objectValue ValueMap
}

// ValueType indicates which JSON type is contained in a Value.
type ValueType int

// String returns the name of the value type.
func (t ValueType) String() string {
	switch t {
	case NullType:
		return nullAsJSON
	case BoolType:
		return "bool"
	case NumberType:
		return "number"
	case StringType:
		return "string"
	case ArrayType:
		return "array"
	case ObjectType:
		return "object"
	case RawType:
		return "raw"
	default:
		return "unknown"
	}
}

// Null creates a null Value.
func Null() Value {
	return Value{valueType: NullType}
}

// Bool creates a boolean Value.
func Bool(value bool) Value {
	return Value{valueType: BoolType, boolValue: value}
}

// Int creates a numeric Value from an integer.
//
// Note that all numbers are represented internally as the same type (float64), so Int(2) is
// exactly equal to Float64(2).
func Int(value int) Value {
	return Float64(float64(value))
}

// Float64 creates a numeric Value from a float64.
func Float64(value float64) Value {
	return Value{valueType: NumberType, numberValue: value}
}

// String creates a string Value.
func String(value string) Value {
	return Value{valueType: StringType, stringValue: value}
}

// Raw creates an unparsed JSON Value.
//
// This constructor will store the json.RawMessage value as-is without syntax validation, and will set
// the type of the Value to ldvalue.RawType. That means, for instance, that ldvalue.Raw(json.RawMessage("true"))
// is not logically equivalent to ldvalue.Bool(true), even though they will produce the same result if
// they are re-encoded to JSON.
//
// That also means that if you pass malformed data that is not valid JSON, you will get malformed data if
// it is re-encoded to JSON. It is the caller's responsibility to make sure the json.RawMessage really is
// valid JSON. However, since it is easy to mistakenly write json.RawMessage(nil) (a zero-length value)
// when what is really meant is a JSON null value, the ldvalue.Value JSON encoder will detect any use of
// json.RawMessage(nil) or json.RawMessage("") and transparently convert it to "null" when it is
// being encoded to JSON.
func Raw(value json.RawMessage) Value {
	return Value{valueType: RawType, stringValue: string(value)}
}

// CopyArbitraryValue creates a Value from an arbitrary interface{} value of any type.
//
// If the value is nil, a boolean, an integer, a floating-point number, or a string, it becomes the
// corresponding JSON primitive value type. If it is a slice of values ([]interface{} or
// []Value), it is deep-copied to an array value. If it is a map of strings to values
// (map[string]interface{} or map[string]Value), it is deep-copied to an object value.
//
// If it is a pointer to any of the above types, then it is dereferenced and treated the same as above,
// unless the pointer is nil, in which case it becomes Null().
//
// For all other types, the value is marshaled to JSON and then converted to the corresponding
// Value type (unless marshalling returns an error, in which case it becomes Null()). This is
// somewhat inefficient since it involves parsing the marshaled JSON.
func CopyArbitraryValue(valueAsInterface interface{}) Value { //nolint:gocyclo // yes, we know it's a long function
	if valueAsInterface == nil {
		return Null()
		// Note that an interface value can be nil in two ways: nil with no type at all, which is this case,
		// or a nil pointer of some specific pointer type, which we have to check for separately below.
	}
	switch o := valueAsInterface.(type) {
	case Value:
		return o
	case *Value:
		if o == nil {
			return Null()
		}
		return *o
	case OptionalString:
		return o.AsValue()
	case *OptionalString:
		if o == nil {
			return Null()
		}
		return o.AsValue()
	case bool:
		return Bool(o)
	case *bool:
		if o == nil {
			return Null()
		}
		return Bool(*o)
	case int8:
		return Float64(float64(o))
	case *int8:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case uint8:
		return Float64(float64(o))
	case *uint8:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case int16:
		return Float64(float64(o))
	case *int16:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case uint16:
		return Float64(float64(o))
	case *uint16:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case int:
		return Float64(float64(o))
	case *int:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case uint:
		return Float64(float64(o))
	case *uint:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case int32:
		return Float64(float64(o))
	case *int32:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case uint32:
		return Float64(float64(o))
	case *uint32:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case float32:
		return Float64(float64(o))
	case *float32:
		if o == nil {
			return Null()
		}
		return Float64(float64(*o))
	case float64:
		return Float64(o)
	case *float64:
		if o == nil {
			return Null()
		}
		return Float64(*o)
	case string:
		return String(o)
	case *string:
		if o == nil {
			return Null()
		}
		return String(*o)
	case []interface{}:
		return copyArbitraryValueArray(o)
	case *[]interface{}:
		if o == nil {
			return Null()
		}
		return copyArbitraryValueArray(*o)
	case []Value:
		return ArrayOf(o...)
	case *[]Value:
		if o == nil {
			return Null()
		}
		return ArrayOf((*o)...)
	case map[string]interface{}:
		return copyArbitraryValueMap(o)
	case *map[string]interface{}:
		if o == nil {
			return Null()
		}
		return copyArbitraryValueMap(*o)
	case map[string]Value:
		return CopyObject(o)
	case *map[string]Value:
		if o == nil {
			return Null()
		}
		return CopyObject(*o)
	case json.RawMessage:
		return Raw(o)
	case *json.RawMessage:
		if o == nil {
			return Null()
		}
		return Raw(*o)
	default:
		jsonBytes, err := json.Marshal(valueAsInterface)
		if err == nil {
			var ret Value
			err = json.Unmarshal(jsonBytes, &ret)
			if err == nil {
				return ret
			}
		}
		return Null()
	}
}

func copyArbitraryValueArray(o []interface{}) Value {
	return Value{valueType: ArrayType, arrayValue: CopyArbitraryValueArray(o)}
}

func copyArbitraryValueMap(o map[string]interface{}) Value {
	return Value{valueType: ObjectType, objectValue: CopyArbitraryValueMap(o)}
}

// Type returns the ValueType of the Value.
func (v Value) Type() ValueType {
	return v.valueType
}

// IsNull returns true if the Value is a null.
func (v Value) IsNull() bool {
	return v.valueType == NullType
}

// IsDefined returns true if the Value is anything other than null.
//
// This is exactly equivalent to !v.IsNull(), but is provided as a separate method for consistency
// with other types that have an IsDefined() method.
func (v Value) IsDefined() bool {
	return !v.IsNull()
}

// IsBool returns true if the Value is a boolean.
func (v Value) IsBool() bool {
	return v.valueType == BoolType
}

// IsNumber returns true if the Value is numeric.
func (v Value) IsNumber() bool {
	return v.valueType == NumberType
}

// IsInt returns true if the Value is an integer.
//
// JSON does not have separate types for integer and floating-point values; they are both just numbers.
// IsInt returns true if and only if the actual numeric value has no fractional component, so
// Int(2).IsInt() and Float64(2.0).IsInt() are both true.
func (v Value) IsInt() bool {
	if v.valueType == NumberType {
		return v.numberValue == float64(int(v.numberValue))
	}
	return false
}

// IsString returns true if the Value is a string.
func (v Value) IsString() bool {
	return v.valueType == StringType
}

// BoolValue returns the Value as a boolean.
//
// If the Value is not a boolean, it returns false.
func (v Value) BoolValue() bool {
	return v.valueType == BoolType && v.boolValue
}

// IntValue returns the value as an int.
//
// If the Value is not numeric, it returns zero. If the value is a number but not an integer, it is
// rounded toward zero (truncated).
func (v Value) IntValue() int {
	if v.valueType == NumberType {
		return int(v.numberValue)
	}
	return 0
}

// Float64Value returns the value as a float64.
//
// If the Value is not numeric, it returns zero.
func (v Value) Float64Value() float64 {
	if v.valueType == NumberType {
		return v.numberValue
	}
	return 0
}

// StringValue returns the value as a string.
//
// If the value is not a string, it returns an empty string.
//
// This is different from String(), which returns a concise string representation of any value type.
func (v Value) StringValue() string {
	if v.valueType == StringType {
		return v.stringValue
	}
	return ""
}

// AsOptionalString converts the value to the OptionalString type, which contains either a string
// value or nothing if the original value was not a string.
func (v Value) AsOptionalString() OptionalString {
	if v.valueType == StringType {
		return NewOptionalString(v.stringValue)
	}
	return OptionalString{}
}

// AsRaw returns the value as a json.RawMessage.
//
// If the value was originally created from a RawMessage, it returns the same value. For all other
// values, it converts the value to its JSON representation and returns that representation.
//
// Note that the ldvalue.Raw(json.RawMessage) constructor does not do any syntax validation, so
// if you create a Value from a malformed string such as ldvalue.Raw(json.RawMessage("{{{")), you
// will get back the same string from AsRaw().
func (v Value) AsRaw() json.RawMessage {
	if v.valueType == RawType {
		return json.RawMessage(v.stringValue)
	}
	bytes, err := json.Marshal(v)
	if err == nil {
		return json.RawMessage(bytes)
	}
	return nil
}

// AsArbitraryValue returns the value in its simplest Go representation, typed as interface{}.
//
// This is nil for a null value; for primitive types, it is bool, float64, or string (all numbers
// are represented as float64 because that is Go's default when parsing from JSON).
//
// Arrays and objects are represented as []interface{} and map[string]interface{}. They are
// deep-copied, which preserves immutability of the Value but may be an expensive operation.
// To examine array and object values without copying the whole data structure, use getter
// methods: Count, Keys, GetByIndex, TryGetByIndex, GetByKey, TryGetByKey.
func (v Value) AsArbitraryValue() interface{} {
	switch v.valueType {
	case NullType:
		return nil
	case BoolType:
		return v.boolValue
	case NumberType:
		return v.numberValue
	case StringType:
		return v.stringValue
	case ArrayType:
		return v.arrayValue.AsArbitraryValueSlice()
	case ObjectType:
		return v.objectValue.AsArbitraryValueMap()
	case RawType:
		return v.AsRaw()
	}
	return nil // should not be possible
}

// String converts the value to a string representation, equivalent to JSONString().
//
// This is different from StringValue, which returns the actual string for a string value or an empty
// string for anything else. For instance, Int(2).StringValue() returns "2" and String("x").StringValue()
// returns "\"x\"", whereas Int(2).AsString() returns "" and String("x").AsString() returns
// "x".
//
// This method is provided because it is common to use the Stringer interface as a quick way to
// summarize the contents of a value. The simplest way to do so in this case is to use the JSON
// representation.
func (v Value) String() string {
	return v.JSONString()
}

// Equal tests whether this Value is equal to another, in both type and value.
//
// For arrays and objects, this is a deep equality test. This method behaves the same as
// reflect.DeepEqual, but is slightly more efficient.
func (v Value) Equal(other Value) bool {
	if v.valueType == other.valueType {
		switch v.valueType {
		case NullType:
			return true
		case BoolType:
			return v.boolValue == other.boolValue
		case NumberType:
			return v.numberValue == other.numberValue
		case StringType, RawType:
			return v.stringValue == other.stringValue
		case ArrayType:
			return v.arrayValue.Equal(other.arrayValue)
		case ObjectType:
			return v.objectValue.Equal(other.objectValue)
		}
	}
	return false
}

// AsPointer returns either a pointer to a copy of this Value, or nil if it is a null value.
//
// This may be desirable if you are serializing a struct that contains a Value, and you want
// that field to be completely omitted if the Value is null; since the "omitempty" tag only
// works for pointers, you can declare the field as a *Value like so:
//
//     type MyJsonStruct struct {
//         AnOptionalField *Value `json:"anOptionalField,omitempty"`
//     }
//     s := MyJsonStruct{AnOptionalField: someValue.AsPointer()}
func (v Value) AsPointer() *Value {
	if v.IsNull() {
		return nil
	}
	return &v
}
