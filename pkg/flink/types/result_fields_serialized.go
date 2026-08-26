package types

import (
	"math"
	"strconv"
)

// ToSerializedValue renders a result field as a value that carries its SQL type into
// `-o json` and `-o yaml`, so a consumer reads a number as a number and a NULL as null
// instead of re-parsing everything out of strings.
//
// This is deliberately not ToSDKType. That one produces the gateway's wire shape, where
// every atom is a string and a MAP is a list of pairs — correct for talking to the API,
// wrong for handing to a script.
//
// Two constraints shape the choices below:
//
//   - Only native Go types are used. json.Number renders as a bare literal through
//     encoding/json but as a quoted string through yaml.v3, which would make the two
//     output formats disagree.
//   - A value that cannot be represented natively keeps the exact text the gateway sent
//     rather than becoming a zero. Losing precision or inventing a plausible number is
//     worse than a string a caller can see and handle.

func (f AtomicStatementResultField) ToSerializedValue() any {
	// A NULL arrives as Type Null carrying the literal text "NULL", which is otherwise
	// indistinguishable from a VARCHAR containing that word.
	if f.Type == Null {
		return nil
	}

	switch f.Type {
	case Boolean:
		if value, err := strconv.ParseBool(f.Value); err == nil {
			return value
		}
	case Tinyint, Smallint, Integer:
		// These top out at 2^31-1, well inside the range a float64 represents exactly,
		// so every consumer reads them back as the number that was sent. BIGINT is not
		// in this list — see below.
		if value, err := strconv.ParseInt(f.Value, 10, 64); err == nil {
			return value
		}
	case Float, Double:
		// NaN and ±Inf are legal in a DOUBLE column but encoding/json refuses them, and
		// failing there would take down a whole result set that drained successfully.
		if value, err := strconv.ParseFloat(f.Value, 64); err == nil && !math.IsNaN(value) && !math.IsInf(value, 0) {
			return value
		}
	}

	// Everything else stays text on purpose:
	//   - BIGINT reaches past 2^53, where a JSON number stops being exact for any
	//     consumer that parses into a float64 — JavaScript and jq both do. Emitting
	//     9007199254740993 as a bare literal silently hands those callers ...992. Go
	//     would keep it exact, but a wire format has to be read by everyone, so the
	//     rule is: any integer type that can leave float64's exact range travels as
	//     text. This is the same reasoning that keeps DECIMAL a string.
	//   - DECIMAL is arbitrary-precision, and float64 would silently round it
	//   - DATE, TIME, TIMESTAMP and INTERVAL have no native JSON form, and reformatting
	//     them would discard the gateway's own rendering
	//   - CHAR, VARCHAR, BINARY and VARBINARY are already text
	return f.Value
}

func (f ArrayStatementResultField) ToSerializedValue() any {
	// Length rather than nil, so an empty array serializes as [] and not null.
	values := make([]any, len(f.Values))
	for idx, value := range f.Values {
		values[idx] = value.ToSerializedValue()
	}
	return values
}

func (f MapStatementResultField) ToSerializedValue() any {
	// A map with textual keys is a JSON object, which is what a consumer expects to see.
	// Anything else cannot become one, so those keep an explicit key/value list instead
	// of the bare positional pairs ToSDKType emits.
	if f.KeyType == Char || f.KeyType == Varchar {
		entries := make(map[string]any, len(f.Entries))
		for _, entry := range f.Entries {
			entries[entry.Key.ToString()] = entry.Value.ToSerializedValue()
		}
		return entries
	}

	entries := make([]any, len(f.Entries))
	for idx, entry := range f.Entries {
		entries[idx] = map[string]any{
			"key":   entry.Key.ToSerializedValue(),
			"value": entry.Value.ToSerializedValue(),
		}
	}
	return entries
}

func (f RowStatementResultField) ToSerializedValue() any {
	// A ROW carries no field names, so it stays positional.
	values := make([]any, len(f.Values))
	for idx, value := range f.Values {
		values[idx] = value.ToSerializedValue()
	}
	return values
}

func (f StructuredTypeStatementResultField) ToSerializedValue() any {
	values := make(map[string]any, len(f.Values))
	for idx, value := range f.Values {
		// FieldNames and Values are built together, but a short FieldNames would panic
		// here, and a partial object beats losing the row.
		if idx >= len(f.FieldNames) {
			break
		}
		values[f.FieldNames[idx]] = value.ToSerializedValue()
	}
	return values
}
