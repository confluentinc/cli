package skillgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMapFlagToJSONSchema_StringType tests mapping string flags to JSON Schema
func TestMapFlagToJSONSchema_StringType(t *testing.T) {
	tests := []struct {
		name     string
		flag     FlagIR
		expected JSONSchema
	}{
		{
			name: "string with no default",
			flag: FlagIR{
				Name: "cluster",
				Type: "string",
			},
			expected: JSONSchema{
				Type: "string",
			},
		},
		{
			name: "string with default value",
			flag: FlagIR{
				Name:    "environment",
				Type:    "string",
				Default: "prod",
				Usage:   "Environment name",
			},
			expected: JSONSchema{
				Type:        "string",
				Default:     "prod",
				Description: "Environment name",
			},
		},
		{
			name: "string with empty default (should omit)",
			flag: FlagIR{
				Name:    "topic",
				Type:    "string",
				Default: "",
				Usage:   "Topic name",
			},
			expected: JSONSchema{
				Type:        "string",
				Description: "Topic name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapFlagToJSONSchema(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMapFlagToJSONSchema_BoolType tests mapping boolean flags to JSON Schema
func TestMapFlagToJSONSchema_BoolType(t *testing.T) {
	tests := []struct {
		name     string
		flag     FlagIR
		expected JSONSchema
	}{
		{
			name: "bool with true default",
			flag: FlagIR{
				Name:    "enable-ssl",
				Type:    "bool",
				Default: "true",
				Usage:   "Enable SSL encryption",
			},
			expected: JSONSchema{
				Type:        "boolean",
				Default:     true,
				Description: "Enable SSL encryption",
			},
		},
		{
			name: "bool with false default (should omit default)",
			flag: FlagIR{
				Name:    "verbose",
				Type:    "bool",
				Default: "false",
			},
			expected: JSONSchema{
				Type: "boolean",
			},
		},
		{
			name: "bool with no default",
			flag: FlagIR{
				Name: "dry-run",
				Type: "bool",
			},
			expected: JSONSchema{
				Type: "boolean",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapFlagToJSONSchema(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMapFlagToJSONSchema_IntegerTypes tests mapping integer flags to JSON Schema
func TestMapFlagToJSONSchema_IntegerTypes(t *testing.T) {
	tests := []struct {
		name     string
		flag     FlagIR
		expected JSONSchema
	}{
		{
			name: "int with non-zero default",
			flag: FlagIR{
				Name:    "partitions",
				Type:    "int",
				Default: "6",
				Usage:   "Number of partitions",
			},
			expected: JSONSchema{
				Type:        "integer",
				Default:     6,
				Description: "Number of partitions",
			},
		},
		{
			name: "int32 with zero default (should omit)",
			flag: FlagIR{
				Name:    "timeout",
				Type:    "int32",
				Default: "0",
			},
			expected: JSONSchema{
				Type: "integer",
			},
		},
		{
			name: "int64 with value",
			flag: FlagIR{
				Name:    "max-bytes",
				Type:    "int64",
				Default: "1048576",
			},
			expected: JSONSchema{
				Type:    "integer",
				Default: 1048576,
			},
		},
		{
			name: "uint with value",
			flag: FlagIR{
				Name:    "replicas",
				Type:    "uint",
				Default: "3",
			},
			expected: JSONSchema{
				Type:    "integer",
				Default: 3,
			},
		},
		{
			name: "uint16 with value",
			flag: FlagIR{
				Name:    "port",
				Type:    "uint16",
				Default: "9092",
			},
			expected: JSONSchema{
				Type:    "integer",
				Default: 9092,
			},
		},
		{
			name: "uint32 with value",
			flag: FlagIR{
				Name:    "buffer-size",
				Type:    "uint32",
				Default: "65536",
			},
			expected: JSONSchema{
				Type:    "integer",
				Default: 65536,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapFlagToJSONSchema(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMapFlagToJSONSchema_StringSliceType tests mapping stringSlice flags to JSON Schema
func TestMapFlagToJSONSchema_StringSliceType(t *testing.T) {
	tests := []struct {
		name     string
		flag     FlagIR
		expected JSONSchema
	}{
		{
			name: "stringSlice basic",
			flag: FlagIR{
				Name:  "topics",
				Type:  "stringSlice",
				Usage: "List of topic names",
			},
			expected: JSONSchema{
				Type:        "array",
				Description: "List of topic names",
				Items: &JSONSchema{
					Type: "string",
				},
			},
		},
		{
			name: "stringSlice with no usage",
			flag: FlagIR{
				Name: "tags",
				Type: "stringSlice",
			},
			expected: JSONSchema{
				Type: "array",
				Items: &JSONSchema{
					Type: "string",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapFlagToJSONSchema(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMapFlagToJSONSchema_StringToStringType tests mapping stringToString flags to JSON Schema
func TestMapFlagToJSONSchema_StringToStringType(t *testing.T) {
	tests := []struct {
		name     string
		flag     FlagIR
		expected JSONSchema
	}{
		{
			name: "stringToString basic",
			flag: FlagIR{
				Name:  "config",
				Type:  "stringToString",
				Usage: "Configuration key-value pairs",
			},
			expected: JSONSchema{
				Type:        "object",
				Description: "Configuration key-value pairs",
				AdditionalProperties: &JSONSchema{
					Type: "string",
				},
			},
		},
		{
			name: "stringToString with no usage",
			flag: FlagIR{
				Name: "labels",
				Type: "stringToString",
			},
			expected: JSONSchema{
				Type: "object",
				AdditionalProperties: &JSONSchema{
					Type: "string",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapFlagToJSONSchema(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMapFlagToJSONSchema_UnknownType tests fallback for unknown flag types
func TestMapFlagToJSONSchema_UnknownType(t *testing.T) {
	flag := FlagIR{
		Name: "custom",
		Type: "unknownType",
	}

	result := MapFlagToJSONSchema(flag)
	assert.Equal(t, "string", result.Type)
}

// TestMapFlagToJSONSchema_FlagNamePreservation tests that flag names are preserved exactly
func TestMapFlagToJSONSchema_FlagNamePreservation(t *testing.T) {
	tests := []struct {
		name         string
		flagName     string
		expectedName string
	}{
		{
			name:         "hyphenated flag preserved",
			flagName:     "my-flag",
			expectedName: "my-flag",
		},
		{
			name:         "single word preserved",
			flagName:     "cluster",
			expectedName: "cluster",
		},
		{
			name:         "multi-hyphenated preserved",
			flagName:     "service-account-id",
			expectedName: "service-account-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := FlagIR{
				Name: tt.flagName,
				Type: "string",
			}
			// The flag name should be preserved in the schema mapping
			// This will be verified when we create the full InputSchema
			result := MapFlagToJSONSchema(flag)
			// At this level, we're just testing the mapping function
			// Full integration test will verify property key preservation
			assert.NotNil(t, result)
		})
	}
}

// TestBuildInputSchema tests building a complete InputSchema from multiple flags
func TestBuildInputSchema(t *testing.T) {
	tests := []struct {
		name     string
		flags    []FlagIR
		expected InputSchema
	}{
		{
			name: "single required flag",
			flags: []FlagIR{
				{
					Name:     "cluster",
					Type:     "string",
					Required: true,
					Usage:    "Cluster ID",
				},
			},
			expected: InputSchema{
				Type: "object",
				Properties: map[string]JSONSchema{
					"cluster": {
						Type:        "string",
						Description: "Cluster ID",
					},
				},
				Required: []string{"cluster"},
			},
		},
		{
			name: "multiple flags with mixed required",
			flags: []FlagIR{
				{
					Name:     "cluster",
					Type:     "string",
					Required: true,
					Usage:    "Cluster ID",
				},
				{
					Name:     "partitions",
					Type:     "int",
					Required: false,
					Default:  "6",
					Usage:    "Number of partitions",
				},
				{
					Name:     "enable-ssl",
					Type:     "bool",
					Required: false,
					Default:  "true",
				},
			},
			expected: InputSchema{
				Type: "object",
				Properties: map[string]JSONSchema{
					"cluster": {
						Type:        "string",
						Description: "Cluster ID",
					},
					"partitions": {
						Type:        "integer",
						Default:     6,
						Description: "Number of partitions",
					},
					"enable-ssl": {
						Type:    "boolean",
						Default: true,
					},
				},
				Required: []string{"cluster"},
			},
		},
		{
			name: "flag name preservation in properties",
			flags: []FlagIR{
				{
					Name: "my-flag",
					Type: "string",
				},
				{
					Name: "service-account-id",
					Type: "string",
				},
			},
			expected: InputSchema{
				Type: "object",
				Properties: map[string]JSONSchema{
					"my-flag": {
						Type: "string",
					},
					"service-account-id": {
						Type: "string",
					},
				},
				Required: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildInputSchema(tt.flags)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Properties, result.Properties)
			assert.ElementsMatch(t, tt.expected.Required, result.Required)
		})
	}
}

// TestMapFlagToJSONSchema_InvalidDefaults tests handling of invalid default values
func TestMapFlagToJSONSchema_InvalidDefaults(t *testing.T) {
	tests := []struct {
		name     string
		flag     FlagIR
		expected JSONSchema
	}{
		{
			name: "invalid bool default",
			flag: FlagIR{
				Name:    "flag",
				Type:    "bool",
				Default: "not-a-bool",
			},
			expected: JSONSchema{
				Type: "boolean",
				// Invalid default should be omitted
			},
		},
		{
			name: "invalid int default",
			flag: FlagIR{
				Name:    "count",
				Type:    "int",
				Default: "not-a-number",
			},
			expected: JSONSchema{
				Type: "integer",
				// Invalid default should be omitted
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapFlagToJSONSchema(tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestToolStructure tests that Tool struct can be properly created
func TestToolStructure(t *testing.T) {
	tool := Tool{
		Name:        "confluent-kafka-topic-list",
		Title:       "List Kafka Topics",
		Description: "List all Kafka topics in a cluster",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]JSONSchema{
				"cluster": {
					Type:        "string",
					Description: "Cluster ID",
				},
			},
			Required: []string{"cluster"},
		},
		Annotations: Annotations{
			Audience:     "customers",
			Priority:     "high",
			LastModified: "2026-03-09",
		},
	}

	require.NotNil(t, tool)
	assert.Equal(t, "confluent-kafka-topic-list", tool.Name)
	assert.Equal(t, "List Kafka Topics", tool.Title)
	assert.Equal(t, "customers", tool.Annotations.Audience)
	assert.Len(t, tool.InputSchema.Required, 1)
}
