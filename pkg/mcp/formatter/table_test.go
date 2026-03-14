package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Additional tests for BuildMarkdownTable focusing on Title Case headers and markdown escaping
func TestBuildMarkdownTable_TitleCase(t *testing.T) {
	t.Run("snake_case headers converted to Title Case", func(t *testing.T) {
		items := []map[string]interface{}{
			{"display_name": "Production Cluster", "environment_id": "env-123"},
		}
		result := BuildMarkdownTable(items)

		// "display_name" -> "Display Name", "environment_id" -> "Environment Id"
		assert.Contains(t, result, "| Display Name | Environment Id |")
	})

	t.Run("single word headers capitalized", func(t *testing.T) {
		items := []map[string]interface{}{
			{"id": "lkc-123", "name": "prod"},
		}
		result := BuildMarkdownTable(items)

		// "id" -> "Id", "name" -> "Name"
		assert.Contains(t, result, "| Id | Name |")
	})
}

func TestBuildMarkdownTable_Escaping(t *testing.T) {
	t.Run("markdown special characters are escaped", func(t *testing.T) {
		items := []map[string]interface{}{
			{"name": "test|pipe", "desc": "has *asterisk* and _underscore_"},
		}
		result := BuildMarkdownTable(items)

		// Pipe, asterisk, underscore should be escaped
		assert.Contains(t, result, "test\\|pipe")
		assert.Contains(t, result, "has \\*asterisk\\* and \\_underscore\\_")
	})
}

func TestBuildMarkdownTable_MissingValues(t *testing.T) {
	t.Run("missing values in some rows show empty cells", func(t *testing.T) {
		items := []map[string]interface{}{
			{"id": "lkc-1", "name": "prod"},
			{"id": "lkc-2"}, // missing name
		}
		result := BuildMarkdownTable(items)

		// Second row should have empty cell for name
		lines := splitLines(result)
		// After Title Case: "Id" and "Name"
		assert.Contains(t, lines[0], "| Id | Name |")
		assert.Contains(t, lines[2], "| lkc-1 | prod |")
		assert.Contains(t, lines[3], "| lkc-2 |  |")
	})
}

func TestFormatHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase word",
			input:    "id",
			expected: "Id",
		},
		{
			name:     "snake_case to Title Case",
			input:    "display_name",
			expected: "Display Name",
		},
		{
			name:     "multiple underscores",
			input:    "environment_id",
			expected: "Environment Id",
		},
		{
			name:     "already capitalized",
			input:    "Name",
			expected: "Name",
		},
		{
			name:     "three word snake_case",
			input:    "kafka_cluster_id",
			expected: "Kafka Cluster Id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHeader(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no special characters",
			input:    "simple text",
			expected: "simple text",
		},
		{
			name:     "pipe character",
			input:    "col1|col2",
			expected: "col1\\|col2",
		},
		{
			name:     "asterisk",
			input:    "bold *text*",
			expected: "bold \\*text\\*",
		},
		{
			name:     "underscore",
			input:    "italic _text_",
			expected: "italic \\_text\\_",
		},
		{
			name:     "multiple special characters",
			input:    "test|*_combined",
			expected: "test\\|\\*\\_combined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMarkdown(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to split result into lines
func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, char := range s {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
