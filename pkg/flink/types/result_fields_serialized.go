package types

import (
	"math"
	"strconv"
)

// ToSerializedValue renders a field as its SQL type for `-o json`/`-o yaml` (a
// number as a number, NULL as null) — unlike ToSDKType, the gateway's string-only wire shape.

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
		// These top out at 2^31-1, well inside float64's exact range. BIGINT isn't
		// here — see below.
		if value, err := strconv.ParseInt(f.Value, 10, 64); err == nil {
			return value
		}
	case Float, Double:
		// NaN/±Inf are legal in a DOUBLE column but encoding/json refuses them; failing
		// here would take down an otherwise-successful drain.
		if value, err := strconv.ParseFloat(f.Value, 64); err == nil && !math.IsNaN(value) && !math.IsInf(value, 0) {
			return value
		}
	}

	// Everything else stays text: BIGINT/DECIMAL would lose precision as float64, and the rest have no native JSON form anyway.
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
	// A map with textual keys becomes a JSON object; anything else keeps an explicit
	// key/value list instead of ToSDKType's bare positional pairs.
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
		// A short FieldNames would panic here; a partial object beats losing the row.
		if idx >= len(f.FieldNames) {
			break
		}
		values[f.FieldNames[idx]] = value.ToSerializedValue()
	}
	return values
}
