package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarizeJSON_Array(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedSummary   string
		expectedCount     int
		expectStructured  bool
	}{
		{
			name:  "simple array with two items",
			input: `[{"id":"lkc-1","name":"prod"},{"id":"lkc-2","name":"dev"}]`,
			expectedSummary: "Found 2 items:",
			expectedCount:   2,
			expectStructured: true,
		},
		{
			name:             "empty array",
			input:            `[]`,
			expectedSummary:  "No items found.",
			expectedCount:    0,
			expectStructured: true,
		},
		{
			name:  "single item array",
			input: `[{"id":"lkc-1","name":"prod"}]`,
			expectedSummary: "Found 1 item:",
			expectedCount:   1,
			expectStructured: true,
		},
		{
			name:  "array with multiple fields",
			input: `[{"id":"lkc-1","name":"prod","status":"UP","region":"us-west-2"}]`,
			expectedSummary: "Found 1 item:",
			expectedCount:   1,
			expectStructured: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, structured := SummarizeJSON(tt.input)

			// Check summary contains expected text
			assert.Contains(t, summary, tt.expectedSummary)

			// Check structured data
			if tt.expectStructured {
				require.NotNil(t, structured)
				items, ok := structured.([]map[string]interface{})
				require.True(t, ok, "structured should be []map[string]interface{}")
				assert.Equal(t, tt.expectedCount, len(items))
			}
		})
	}
}

func TestSummarizeJSON_Object(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedInSummary []string
		expectStructured bool
	}{
		{
			name:  "object with id and name",
			input: `{"id":"lkc-123","name":"prod","status":"UP"}`,
			expectedInSummary: []string{
				"prod (lkc-123):",
				"• id: lkc-123",
				"• name: prod",
				"• status: UP",
			},
			expectStructured: true,
		},
		{
			name:  "object with id only",
			input: `{"id":"lkc-123","status":"UP"}`,
			expectedInSummary: []string{
				"lkc-123:",
				"• id: lkc-123",
				"• status: UP",
			},
			expectStructured: true,
		},
		{
			name:  "object with display_name",
			input: `{"id":"env-123","display_name":"Production Environment"}`,
			expectedInSummary: []string{
				"Production Environment (env-123):",
				"• display_name: Production Environment",
				"• id: env-123",
			},
			expectStructured: true,
		},
		{
			name:  "object with label field",
			input: `{"id":"res-123","label":"My Resource"}`,
			expectedInSummary: []string{
				"My Resource (res-123):",
				"• id: res-123",
				"• label: My Resource",
			},
			expectStructured: true,
		},
		{
			name:  "object with title field",
			input: `{"id":"doc-123","title":"Documentation"}`,
			expectedInSummary: []string{
				"Documentation (doc-123):",
				"• id: doc-123",
				"• title: Documentation",
			},
			expectStructured: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, structured := SummarizeJSON(tt.input)

			// Check all expected strings are in summary
			for _, expected := range tt.expectedInSummary {
				assert.Contains(t, summary, expected, "summary should contain: %s", expected)
			}

			// Check structured data
			if tt.expectStructured {
				require.NotNil(t, structured)
				item, ok := structured.(map[string]interface{})
				require.True(t, ok, "structured should be map[string]interface{}")
				assert.NotEmpty(t, item)
			}
		})
	}
}

func TestSummarizeJSON_NonJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "plain text output",
			expected: "plain text output",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "error message",
			input:    "Error: command failed",
			expected: "Error: command failed",
		},
		{
			name:     "malformed JSON",
			input:    `{"incomplete": `,
			expected: `{"incomplete": `,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, structured := SummarizeJSON(tt.input)

			assert.Equal(t, tt.expected, summary)
			assert.Nil(t, structured)
		})
	}
}

func TestExtractID(t *testing.T) {
	tests := []struct {
		name     string
		item     map[string]interface{}
		expected string
	}{
		{
			name:     "has id field",
			item:     map[string]interface{}{"id": "lkc-123", "name": "prod"},
			expected: "lkc-123",
		},
		{
			name:     "no id field",
			item:     map[string]interface{}{"name": "prod"},
			expected: "",
		},
		{
			name:     "empty map",
			item:     map[string]interface{}{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractID(tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractName(t *testing.T) {
	tests := []struct {
		name     string
		item     map[string]interface{}
		expected string
	}{
		{
			name:     "has name field",
			item:     map[string]interface{}{"id": "lkc-123", "name": "prod"},
			expected: "prod",
		},
		{
			name:     "has display_name field",
			item:     map[string]interface{}{"id": "env-123", "display_name": "Production"},
			expected: "Production",
		},
		{
			name:     "has label field",
			item:     map[string]interface{}{"id": "res-123", "label": "My Resource"},
			expected: "My Resource",
		},
		{
			name:     "has title field",
			item:     map[string]interface{}{"id": "doc-123", "title": "Documentation"},
			expected: "Documentation",
		},
		{
			name:     "name takes precedence over display_name",
			item:     map[string]interface{}{"name": "prod", "display_name": "Production"},
			expected: "prod",
		},
		{
			name:     "no name field",
			item:     map[string]interface{}{"id": "lkc-123", "status": "UP"},
			expected: "",
		},
		{
			name:     "empty map",
			item:     map[string]interface{}{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractName(tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatSingle(t *testing.T) {
	tests := []struct {
		name     string
		item     map[string]interface{}
		expected []string
	}{
		{
			name: "with name and id",
			item: map[string]interface{}{"id": "lkc-123", "name": "prod", "status": "UP"},
			expected: []string{
				"prod (lkc-123):",
				"• id: lkc-123",
				"• name: prod",
				"• status: UP",
			},
		},
		{
			name: "id only",
			item: map[string]interface{}{"id": "lkc-123", "status": "UP"},
			expected: []string{
				"lkc-123:",
				"• id: lkc-123",
				"• status: UP",
			},
		},
		{
			name: "fields sorted alphabetically",
			item: map[string]interface{}{"zebra": "z", "alpha": "a", "beta": "b"},
			expected: []string{
				"• alpha: a",
				"• beta: b",
				"• zebra: z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSingle(tt.item)

			for _, expectedLine := range tt.expected {
				assert.Contains(t, result, expectedLine)
			}
		})
	}
}

func TestFormatList(t *testing.T) {
	tests := []struct {
		name     string
		items    []map[string]interface{}
		expected []string
	}{
		{
			name: "two items",
			items: []map[string]interface{}{
				{"id": "lkc-1", "name": "prod"},
				{"id": "lkc-2", "name": "dev"},
			},
			expected: []string{
				"Found 2 items:",
			},
		},
		{
			name: "one item",
			items: []map[string]interface{}{
				{"id": "lkc-1", "name": "prod"},
			},
			expected: []string{
				"Found 1 item:",
			},
		},
		{
			name:  "empty list",
			items: []map[string]interface{}{},
			expected: []string{
				"No items found.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatList(tt.items)

			for _, expectedLine := range tt.expected {
				assert.Contains(t, result, expectedLine)
			}
		})
	}
}

func TestBuildMarkdownTable(t *testing.T) {
	tests := []struct {
		name     string
		items    []map[string]interface{}
		expected []string
	}{
		{
			name: "simple table",
			items: []map[string]interface{}{
				{"id": "lkc-1", "name": "prod"},
				{"id": "lkc-2", "name": "dev"},
			},
			expected: []string{
				"| Id | Name |", // Title Case headers
				"| --- | --- |",
				"| lkc-1 | prod |",
				"| lkc-2 | dev |",
			},
		},
		{
			name: "empty items",
			items: []map[string]interface{}{},
			expected: []string{
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildMarkdownTable(tt.items)

			for _, expectedLine := range tt.expected {
				if expectedLine != "" {
					assert.Contains(t, result, expectedLine)
				}
			}
		})
	}
}
