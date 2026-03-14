package skillgen

import (
	"strconv"
)

// MapFlagToJSONSchema converts a FlagIR to a JSONSchema parameter definition.
// It handles type mapping, default value parsing, and description preservation.
//
// Type mappings:
//   - string → {"type": "string"}
//   - bool → {"type": "boolean", "default": true/false}
//   - int/int32/int64/uint/uint16/uint32 → {"type": "integer", "default": N}
//   - stringSlice → {"type": "array", "items": {"type": "string"}}
//   - stringToString → {"type": "object", "additionalProperties": {"type": "string"}}
//   - unknown → {"type": "string"} (fallback)
//
// Default value handling:
//   - Empty string defaults are omitted
//   - Zero-value defaults (0, false) are omitted
//   - Non-zero defaults are parsed to correct JSON type
//   - Invalid defaults are omitted (with silent fallback)
//
// Flag names are preserved exactly as-is (no transformation).
func MapFlagToJSONSchema(flag FlagIR) JSONSchema {
	schema := JSONSchema{}

	// Map flag type to JSON Schema type
	switch flag.Type {
	case "string":
		schema.Type = "string"
		// Include default only if non-empty
		if flag.Default != "" {
			schema.Default = flag.Default
		}

	case "bool":
		schema.Type = "boolean"
		// Parse bool default — pflag always provides valid bool strings
		// ("true"/"false"), so parse errors here indicate unexpected input
		// and are safely omitted (zero value "false" is also omitted).
		if flag.Default != "" {
			if boolVal, err := strconv.ParseBool(flag.Default); err == nil {
				if boolVal {
					schema.Default = boolVal
				}
			}
		}

	case "int", "int32", "int64", "uint", "uint16", "uint32":
		schema.Type = "integer"
		// Parse integer default — pflag always provides valid integer strings,
		// so parse errors here indicate unexpected input and are safely
		// omitted (zero value 0 is also omitted).
		if flag.Default != "" {
			if intVal, err := strconv.Atoi(flag.Default); err == nil {
				if intVal != 0 {
					schema.Default = intVal
				}
			}
		}

	case "stringSlice":
		schema.Type = "array"
		schema.Items = &JSONSchema{
			Type: "string",
		}

	case "stringToString":
		schema.Type = "object"
		schema.AdditionalProperties = &JSONSchema{
			Type: "string",
		}

	default:
		// Unknown types fall back to string
		schema.Type = "string"
		if flag.Default != "" {
			schema.Default = flag.Default
		}
	}

	// Include description from flag usage
	if flag.Usage != "" {
		schema.Description = flag.Usage
	}

	return schema
}

// BuildInputSchema constructs a complete InputSchema from a list of flags.
// It creates the properties map and builds the required array.
//
// Flag names are used directly as property keys (no transformation).
// Required flags are collected into the Required array.
func BuildInputSchema(flags []FlagIR) InputSchema {
	schema := InputSchema{
		Type:       "object",
		Properties: make(map[string]JSONSchema, len(flags)),
		Required:   make([]string, 0, len(flags)),
	}

	for _, flag := range flags {
		// Map flag to JSON Schema and store with flag name as key
		// CRITICAL: Use flag.Name directly - no transformation
		schema.Properties[flag.Name] = MapFlagToJSONSchema(flag)

		// Collect required flags
		if flag.Required {
			schema.Required = append(schema.Required, flag.Name)
		}
	}

	return schema
}
