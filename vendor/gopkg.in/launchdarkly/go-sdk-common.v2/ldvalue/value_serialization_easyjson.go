//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package ldvalue

import (
	"github.com/mailru/easyjson/jlexer"
	ej_jwriter "github.com/mailru/easyjson/jwriter"
)

// This conditionally-compiled file provides custom marshal/unmarshal functions for ldvalue types
// in EasyJSON.
//
// EasyJSON's code generator does recognize the same MarshalJSON and UnmarshalJSON methods used by
// encoding/json, and will call them if present. But this mechanism is inefficient: when marshaling
// it requires the allocation of intermediate byte slices, and when unmarshaling it causes the
// JSON object to be parsed twice. It is preferable to have our marshal/unmarshal methods write to
// and read from the EasyJSON Writer/Lexer directly. Also, since deserialization is a high-traffic
// path in some LaunchDarkly code on the service side, the extra overhead of the go-jsonstream
// abstraction is undesirable.

// For more information, see: https://gopkg.in/launchdarkly/go-jsonstream.v1

func (v Value) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	switch v.valueType {
	case NullType:
		writer.Raw(nullAsJSONBytes, nil)
	case BoolType:
		writer.Bool(v.boolValue)
	case NumberType:
		writer.Float64(v.numberValue)
	case StringType:
		writer.String(v.stringValue)
	case ArrayType:
		v.arrayValue.MarshalEasyJSON(writer)
	case ObjectType:
		v.objectValue.MarshalEasyJSON(writer)
	case RawType:
		if len(v.stringValue) == 0 {
			writer.Raw(nullAsJSONBytes, nil) // see Value.MarshalJSON
		} else {
			writer.RawString(v.stringValue)
		}
	}
}

func (v *Value) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsDelim('[') {
		var va ValueArray
		va.UnmarshalEasyJSON(lexer)
		*v = Value{valueType: ArrayType, arrayValue: va}
	} else if lexer.IsDelim('{') {
		var vm ValueMap
		vm.UnmarshalEasyJSON(lexer)
		*v = Value{valueType: ObjectType, objectValue: vm}
	} else {
		*v = CopyArbitraryValue(lexer.Interface())
	}
}

func (v OptionalBool) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if v.hasValue {
		writer.Bool(v.value)
	} else {
		writer.Raw(nullAsJSONBytes, nil)
	}
}

func (v *OptionalBool) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*v = OptionalBool{}
		return
	}
	v.hasValue = true
	v.value = lexer.Bool()
}

func (v OptionalInt) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if v.hasValue {
		writer.Int(v.value)
	} else {
		writer.Raw(nullAsJSONBytes, nil)
	}
}

func (v *OptionalInt) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*v = OptionalInt{}
		return
	}
	v.hasValue = true
	v.value = lexer.Int()
}

func (v OptionalString) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if v.hasValue {
		writer.String(v.value)
	} else {
		writer.Raw(nullAsJSONBytes, nil)
	}
}

func (v *OptionalString) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*v = OptionalString{}
		return
	}
	v.hasValue = true
	v.value = lexer.String()
}

func (v ValueArray) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if v.data == nil {
		writer.Raw(nullAsJSONBytes, nil)
		return
	}
	writer.RawByte('[')
	for i, value := range v.data {
		if i != 0 {
			writer.RawByte(',')
		}
		value.MarshalEasyJSON(writer)
	}
	writer.RawByte(']')
}

func (v *ValueArray) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*v = ValueArray{}
		return
	}
	lexer.Delim('[')
	v.data = make([]Value, 0, 4)
	for !lexer.IsDelim(']') {
		var value Value
		value.UnmarshalEasyJSON(lexer)
		v.data = append(v.data, value)
		lexer.WantComma()
	}
	lexer.Delim(']')
}

func (v ValueMap) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if v.data == nil {
		writer.Raw(nullAsJSONBytes, nil)
		return
	}
	writer.RawByte('{')
	first := true
	for key, value := range v.data {
		if !first {
			writer.RawByte(',')
		}
		first = false
		writer.String(key)
		writer.RawByte(':')
		value.MarshalEasyJSON(writer)
	}
	writer.RawByte('}')
}

func (v *ValueMap) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Null()
		*v = ValueMap{}
		return
	}
	v.data = make(map[string]Value)
	lexer.Delim('{')
	for !lexer.IsDelim('}') {
		key := string(lexer.String())
		lexer.WantColon()
		var value Value
		value.UnmarshalEasyJSON(lexer)
		v.data[key] = value
		lexer.WantComma()
	}
	lexer.Delim('}')
}
