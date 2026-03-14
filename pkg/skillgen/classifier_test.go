package skillgen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInferOperation(t *testing.T) {
	tests := []struct {
		name        string
		commandPath string
		expected    string
	}{
		// Standard CRUD operations
		{
			name:        "list operation",
			commandPath: "confluent kafka topic list",
			expected:    "list",
		},
		{
			name:        "create operation",
			commandPath: "confluent kafka cluster create",
			expected:    "create",
		},
		{
			name:        "delete operation",
			commandPath: "confluent iam service-account delete",
			expected:    "delete",
		},
		{
			name:        "describe operation",
			commandPath: "confluent kafka topic describe",
			expected:    "describe",
		},
		{
			name:        "update operation",
			commandPath: "confluent kafka cluster update",
			expected:    "update",
		},
		{
			name:        "use operation",
			commandPath: "confluent environment use",
			expected:    "use",
		},
		{
			name:        "start operation",
			commandPath: "confluent local kafka start",
			expected:    "start",
		},
		{
			name:        "stop operation",
			commandPath: "confluent local kafka stop",
			expected:    "stop",
		},
		// Edge cases - non-CRUD commands
		{
			name:        "login returns other",
			commandPath: "confluent login",
			expected:    "other",
		},
		{
			name:        "logout returns other",
			commandPath: "confluent logout",
			expected:    "other",
		},
		{
			name:        "version returns other",
			commandPath: "confluent version",
			expected:    "other",
		},
		{
			name:        "shell returns other",
			commandPath: "confluent flink shell",
			expected:    "other",
		},
		{
			name:        "help returns other",
			commandPath: "confluent kafka topic help",
			expected:    "other",
		},
		// Multi-word resources
		{
			name:        "service account create",
			commandPath: "confluent iam service-account create",
			expected:    "create",
		},
		{
			name:        "schema registry schema list",
			commandPath: "confluent schema-registry schema list",
			expected:    "list",
		},
		// Hyphenated verbs (if they exist)
		{
			name:        "hyphenated resource name preserved",
			commandPath: "confluent api-key list",
			expected:    "list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferOperation(tt.commandPath)
			assert.Equal(t, tt.expected, result, "InferOperation(%q) should return %q", tt.commandPath, tt.expected)
		})
	}
}

func TestInferResource(t *testing.T) {
	tests := []struct {
		name        string
		commandPath string
		expected    string
	}{
		// Single-word resources
		{
			name:        "kafka topic",
			commandPath: "confluent kafka topic list",
			expected:    "kafka-topic",
		},
		{
			name:        "kafka cluster",
			commandPath: "confluent kafka cluster create",
			expected:    "kafka-cluster",
		},
		{
			name:        "environment",
			commandPath: "confluent environment use",
			expected:    "environment",
		},
		// Multi-word resources
		{
			name:        "iam service-account",
			commandPath: "confluent iam service-account delete",
			expected:    "iam-service-account",
		},
		{
			name:        "iam user",
			commandPath: "confluent iam user list",
			expected:    "iam-user",
		},
		{
			name:        "schema-registry schema",
			commandPath: "confluent schema-registry schema create",
			expected:    "schema-registry-schema",
		},
		// Edge cases - root commands
		{
			name:        "login has empty resource",
			commandPath: "confluent login",
			expected:    "",
		},
		{
			name:        "logout has empty resource",
			commandPath: "confluent logout",
			expected:    "",
		},
		{
			name:        "version has empty resource",
			commandPath: "confluent version",
			expected:    "",
		},
		// Complex paths
		{
			name:        "local kafka",
			commandPath: "confluent local kafka start",
			expected:    "local-kafka",
		},
		{
			name:        "flink shell",
			commandPath: "confluent flink shell",
			expected:    "flink",
		},
		// API key resource
		{
			name:        "api-key",
			commandPath: "confluent api-key list",
			expected:    "api-key",
		},
		{
			name:        "api-key create",
			commandPath: "confluent api-key create",
			expected:    "api-key",
		},
		// Audit log resource
		{
			name:        "audit-log describe",
			commandPath: "confluent audit-log describe",
			expected:    "audit-log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := InferResource(tt.commandPath)
			assert.Equal(t, tt.expected, result, "InferResource(%q) should return %q", tt.commandPath, tt.expected)
		})
	}
}

func TestClassificationEdgeCases(t *testing.T) {
	t.Run("empty command path", func(t *testing.T) {
		assert.Equal(t, "other", InferOperation(""))
		assert.Equal(t, "", InferResource(""))
	})

	t.Run("single token command", func(t *testing.T) {
		assert.Equal(t, "other", InferOperation("confluent"))
		assert.Equal(t, "", InferResource("confluent"))
	})

	t.Run("two token command", func(t *testing.T) {
		assert.Equal(t, "other", InferOperation("confluent login"))
		assert.Equal(t, "", InferResource("confluent login"))
	})

	t.Run("whitespace handling", func(t *testing.T) {
		assert.Equal(t, "list", InferOperation("  confluent  kafka  topic  list  "))
		assert.Equal(t, "kafka-topic", InferResource("  confluent  kafka  topic  list  "))
	})
}
