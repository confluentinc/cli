package types

import (
	"math"
	"strconv"
)

// ToSerializedValue renders a field as a value that carries its SQL type into
// `-o json`/`-o yaml`, so a number reads as a number and NULL as null.
//
// Not ToSDKType: that produces the gateway's wire shape (every atom a string, MAP
// as pairs) — correct for the API, wrong for a script.
//
// Two constraints: only native Go types are used (json.Number renders as a bare
// literal in encoding/json but quoted in yaml.v3, which would make the formats
// disagree), and a value that can't be represented natively keeps the gateway's
// exact text rather than becoming a zero or an approximation.

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

	// Everything else stays text on purpose:
	//   - BIGINT exceeds float64's exact range (2^53); JS/jq parse JSON numbers as
	//     float64, so a bare literal would silently corrupt large values. Same
	//     reasoning keeps DECIMAL a string.
	//   - DECIMAL is arbitrary-precision; float64 would round it.
	//   - DATE/TIME/TIMESTAMP/INTERVAL have no native JSON form.
	//   - CHAR/VARCHAR/BINARY/VARBINARY are already text.
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
