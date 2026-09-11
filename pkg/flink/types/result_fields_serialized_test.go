package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAtomicToSerializedValue(t *testing.T) {
	tests := []struct {
		name  string
		field AtomicStatementResultField
		want  any
	}{
		{
			// A NULL and a CHAR "NULL" arrive as the same text, told apart only by type.
			name:  "a NULL becomes JSON null",
			field: AtomicStatementResultField{Type: Null, Value: "NULL"},
			want:  nil,
		},
		{
			name:  "a string holding the word NULL stays a string",
			field: AtomicStatementResultField{Type: Varchar, Value: "NULL"},
			want:  "NULL",
		},
		{name: "integer", field: AtomicStatementResultField{Type: Integer, Value: "3065"}, want: int64(3065)},
		{name: "negative integer", field: AtomicStatementResultField{Type: Integer, Value: "-7"}, want: int64(-7)},
		{name: "tinyint", field: AtomicStatementResultField{Type: Tinyint, Value: "12"}, want: int64(12)},
		{name: "smallint", field: AtomicStatementResultField{Type: Smallint, Value: "300"}, want: int64(300)},
		{name: "double", field: AtomicStatementResultField{Type: Double, Value: "52.48"}, want: 52.48},
		{name: "float", field: AtomicStatementResultField{Type: Float, Value: "0.5"}, want: 0.5},
		{name: "boolean true", field: AtomicStatementResultField{Type: Boolean, Value: "true"}, want: true},
		{name: "boolean false", field: AtomicStatementResultField{Type: Boolean, Value: "false"}, want: false},
		{
			// float64 can't represent every BIGINT; JS/jq both parse JSON numbers as float64.
			name:  "bigint stays text so it survives a float64 parser",
			field: AtomicStatementResultField{Type: Bigint, Value: "9007199254740993"},
			want:  "9007199254740993",
		},
		{
			name:  "decimal stays text so precision is not rounded away",
			field: AtomicStatementResultField{Type: Decimal, Value: "1.23"},
			want:  "1.23",
		},
		{
			// encoding/json refuses NaN/±Inf; falling back to text avoids failing the marshal.
			name:  "NaN falls back to text rather than failing the marshal",
			field: AtomicStatementResultField{Type: Double, Value: "NaN"},
			want:  "NaN",
		},
		{
			name:  "infinity falls back to text",
			field: AtomicStatementResultField{Type: Double, Value: "Infinity"},
			want:  "Infinity",
		},
		{
			name:  "a number that does not parse keeps its text instead of becoming zero",
			field: AtomicStatementResultField{Type: Integer, Value: "not-a-number"},
			want:  "not-a-number",
		},
		{name: "timestamp keeps the gateway's rendering", field: AtomicStatementResultField{Type: TimestampWithoutTimeZone, Value: "2026-08-13 10:00:00"}, want: "2026-08-13 10:00:00"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, test.field.ToSerializedValue())
		})
	}
}

// A BIGINT past 2^53 has to survive the round trip a JSON consumer actually performs.
func TestBigintSurvivesAFloat64Parser(t *testing.T) {
	const exact = "9007199254740993"
	field := AtomicStatementResultField{Type: Bigint, Value: exact}

	encoded, err := json.Marshal(map[string]any{"big": field.ToSerializedValue()})
	require.NoError(t, err)

	// json.Unmarshal into `any` decodes every number as float64, which is what a
	// JavaScript client does too.
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(encoded, &decoded))
	require.Equal(t, exact, decoded["big"], "BIGINT lost precision through a float64 parser")
}

func TestArrayToSerializedValue(t *testing.T) {
	field := ArrayStatementResultField{
		ElementType: Integer,
		Values: []StatementResultField{
			AtomicStatementResultField{Type: Integer, Value: "1"},
			AtomicStatementResultField{Type: Null, Value: "NULL"},
		},
	}
	require.Equal(t, []any{int64(1), nil}, field.ToSerializedValue())
}

// An empty array must serialize as [] rather than null.
func TestEmptyArraySerializesAsEmptyList(t *testing.T) {
	field := ArrayStatementResultField{ElementType: Integer}
	encoded, err := json.Marshal(field.ToSerializedValue())
	require.NoError(t, err)
	require.Equal(t, "[]", string(encoded))
}

func TestMapWithTextKeysBecomesAnObject(t *testing.T) {
	field := MapStatementResultField{
		KeyType:   Varchar,
		ValueType: Integer,
		Entries: []MapStatementResultFieldEntry{{
			Key:   AtomicStatementResultField{Type: Varchar, Value: "a"},
			Value: AtomicStatementResultField{Type: Integer, Value: "1"},
		}},
	}
	require.Equal(t, map[string]any{"a": int64(1)}, field.ToSerializedValue())
}

// A JSON object cannot have a non-textual key, so those keep an explicit key/value list.
func TestMapWithNonTextKeysKeepsPairs(t *testing.T) {
	field := MapStatementResultField{
		KeyType:   Integer,
		ValueType: Varchar,
		Entries: []MapStatementResultFieldEntry{{
			Key:   AtomicStatementResultField{Type: Integer, Value: "1"},
			Value: AtomicStatementResultField{Type: Varchar, Value: "a"},
		}},
	}
	require.Equal(t, []any{map[string]any{"key": int64(1), "value": "a"}}, field.ToSerializedValue())
}

func TestRowStaysPositional(t *testing.T) {
	field := RowStatementResultField{
		ElementTypes: []StatementResultFieldType{Integer, Varchar},
		Values: []StatementResultField{
			AtomicStatementResultField{Type: Integer, Value: "1"},
			AtomicStatementResultField{Type: Varchar, Value: "a"},
		},
	}
	require.Equal(t, []any{int64(1), "a"}, field.ToSerializedValue())
}

func TestStructuredTypeUsesFieldNames(t *testing.T) {
	field := StructuredTypeStatementResultField{
		FieldNames: []string{"id", "label"},
		FieldTypes: []StatementResultFieldType{Integer, Varchar},
		Values: []StatementResultField{
			AtomicStatementResultField{Type: Integer, Value: "1"},
			AtomicStatementResultField{Type: Varchar, Value: "a"},
		},
	}
	require.Equal(t, map[string]any{"id": int64(1), "label": "a"}, field.ToSerializedValue())
}

// Fewer names than values must not panic; a partial object beats losing the row.
func TestStructuredTypeWithShortFieldNamesDoesNotPanic(t *testing.T) {
	field := StructuredTypeStatementResultField{
		FieldNames: []string{"id"},
		Values: []StatementResultField{
			AtomicStatementResultField{Type: Integer, Value: "1"},
			AtomicStatementResultField{Type: Varchar, Value: "dropped"},
		},
	}
	require.Equal(t, map[string]any{"id": int64(1)}, field.ToSerializedValue())
}
