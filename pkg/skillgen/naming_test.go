package skillgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuildToolName tests tool name generation for all tier levels
func TestBuildToolName(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		resource  string
		operation string
		tier      string
		expected  string
	}{
		{
			name:      "high tier with resource and operation",
			namespace: "kafka",
			resource:  "topic",
			operation: "list",
			tier:      "high",
			expected:  "confluent_kafka_topic_list",
		},
		{
			name:      "high tier with hyphenated resource",
			namespace: "kafka",
			resource:  "acl-topic",
			operation: "create",
			tier:      "high",
			expected:  "confluent_kafka_acl_topic_create",
		},
		{
			name:      "medium tier with operation only",
			namespace: "schema-registry",
			resource:  "",
			operation: "create",
			tier:      "medium",
			expected:  "confluent_schema_registry_create",
		},
		{
			name:      "medium tier with hyphenated namespace",
			namespace: "schema-registry",
			resource:  "",
			operation: "list",
			tier:      "medium",
			expected:  "confluent_schema_registry_list",
		},
		{
			name:      "low tier namespace only",
			namespace: "plugin",
			resource:  "",
			operation: "",
			tier:      "low",
			expected:  "confluent_plugin",
		},
		{
			name:      "low tier with hyphenated namespace",
			namespace: "api-key",
			resource:  "",
			operation: "",
			tier:      "low",
			expected:  "confluent_api_key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildToolName(tt.namespace, tt.resource, tt.operation, tt.tier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBuildToolName_SpecialCommands tests special handling for login, logout, version, update
func TestBuildToolName_SpecialCommands(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		expected  string
	}{
		{
			name:      "login command",
			namespace: "login",
			expected:  "confluent_login",
		},
		{
			name:      "logout command",
			namespace: "logout",
			expected:  "confluent_logout",
		},
		{
			name:      "version command",
			namespace: "version",
			expected:  "confluent_version",
		},
		{
			name:      "update command",
			namespace: "update",
			expected:  "confluent_update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildToolName(tt.namespace, "", "", "low")
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSanitizeName tests name sanitization (hyphens to underscores)
func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no hyphens",
			input:    "kafka",
			expected: "kafka",
		},
		{
			name:     "single hyphen",
			input:    "api-key",
			expected: "api_key",
		},
		{
			name:     "multiple hyphens",
			input:    "kafka-acl-topic",
			expected: "kafka_acl_topic",
		},
		{
			name:     "mixed underscores and hyphens",
			input:    "schema_registry-config",
			expected: "schema_registry_config",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateToolName tests MCP SEP-986 compliance
func TestValidateToolName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		isValid bool
	}{
		{
			name:    "valid basic name",
			input:   "confluent_kafka_topic_list",
			isValid: true,
		},
		{
			name:    "valid with hyphens",
			input:   "confluent-kafka-topic-list",
			isValid: true,
		},
		{
			name:    "valid with dots",
			input:   "confluent.kafka.topic.list",
			isValid: true,
		},
		{
			name:    "valid with slashes",
			input:   "confluent/kafka/topic/list",
			isValid: true,
		},
		{
			name:    "valid mixed separators",
			input:   "confluent_kafka-topic.list/all",
			isValid: true,
		},
		{
			name:    "empty string invalid",
			input:   "",
			isValid: false,
		},
		{
			name:    "too long (>64 chars)",
			input:   "confluent_kafka_topic_list_with_a_very_long_name_that_exceeds_sixty_four_characters",
			isValid: false,
		},
		{
			name:    "invalid character space",
			input:   "confluent kafka",
			isValid: false,
		},
		{
			name:    "invalid character @",
			input:   "confluent@kafka",
			isValid: false,
		},
		{
			name:    "exactly 64 characters",
			input:   "confluent_kafka_topic_list_operation_with_exactly_64_chars_here",
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateToolName(tt.input)
			assert.Equal(t, tt.isValid, result)
		})
	}
}
