package formatter

import (
	"fmt"
	"sort"
	"strings"
)

// BuildMarkdownTable creates a markdown table from a list of items.
// Headers are converted to Title Case (snake_case → Title Case).
// All columns are left-aligned using "| --- |" separator format.
// Markdown special characters (|, *, _) are escaped in cell values.
// Returns empty string if items is empty.
func BuildMarkdownTable(items []map[string]interface{}) string {
	if len(items) == 0 {
		return ""
	}

	// Extract all unique keys across all items
	keySet := make(map[string]bool)
	for _, item := range items {
		for key := range item {
			keySet[key] = true
		}
	}

	// Sort keys alphabetically for consistency
	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build header row with Title Case formatting
	headerCells := make([]string, len(keys))
	for i, key := range keys {
		headerCells[i] = formatHeader(key)
	}
	lines := []string{
		"| " + strings.Join(headerCells, " | ") + " |",
	}

	// Build separator row (left-aligned: "| --- |")
	separatorCells := make([]string, len(keys))
	for i := range keys {
		separatorCells[i] = "---"
	}
	lines = append(lines, "| "+strings.Join(separatorCells, " | ")+" |")

	// Build data rows
	for _, item := range items {
		dataCells := make([]string, len(keys))
		for i, key := range keys {
			if value, ok := item[key]; ok {
				// Format value and escape markdown characters
				formatted := fmt.Sprintf("%v", value)
				dataCells[i] = escapeMarkdown(formatted)
			} else {
				dataCells[i] = ""
			}
		}
		lines = append(lines, "| "+strings.Join(dataCells, " | ")+" |")
	}

	return strings.Join(lines, "\n")
}

// formatHeader converts snake_case field names to Title Case for table headers.
// Examples:
//   - "id" → "Id"
//   - "display_name" → "Display Name"
//   - "kafka_cluster_id" → "Kafka Cluster Id"
func formatHeader(key string) string {
	// Split on underscores
	words := strings.Split(key, "_")

	// Capitalize each word
	for i, word := range words {
		if len(word) > 0 {
			// Capitalize first letter, keep rest as-is
			words[i] = strings.ToUpper(string(word[0])) + word[1:]
		}
	}

	// Join with spaces
	return strings.Join(words, " ")
}

// escapeMarkdown escapes markdown special characters in table cell values
// to prevent markdown injection and formatting issues.
// Escapes: |, *, _
func escapeMarkdown(value string) string {
	value = strings.ReplaceAll(value, "|", "\\|")
	value = strings.ReplaceAll(value, "*", "\\*")
	value = strings.ReplaceAll(value, "_", "\\_")
	return value
}
