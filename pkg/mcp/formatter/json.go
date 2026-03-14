package formatter

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// SummarizeJSON attempts to parse and format JSON output into human-readable text.
// It detects JSON arrays (lists), objects (single resources), and handles non-JSON passthrough.
//
// Returns:
// - summary: Human-readable formatted text (markdown tables for arrays, bullet lists for objects)
// - structured: Parsed JSON data ([]map[string]interface{} for arrays, map[string]interface{} for objects, nil for non-JSON)
func SummarizeJSON(output string) (summary string, structured interface{}) {
	trimmed := strings.TrimSpace(output)

	// Try to parse as JSON array (list of resources)
	if strings.HasPrefix(trimmed, "[") {
		var items []map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &items); err == nil {
			summary = formatList(items)
			return summary, items
		}
	}

	// Try to parse as JSON object (single resource)
	if strings.HasPrefix(trimmed, "{") {
		var item map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &item); err == nil {
			summary = formatSingle(item)
			return summary, item
		}
	}

	// Not JSON or parsing failed - return as-is
	return output, nil
}

// formatList formats a list of items with a count header and markdown table.
// Empty lists return "No items found."
func formatList(items []map[string]interface{}) string {
	if len(items) == 0 {
		return "No items found."
	}

	// Count header with proper singular/plural
	var header string
	if len(items) == 1 {
		header = "Found 1 item:"
	} else {
		header = fmt.Sprintf("Found %d items:", len(items))
	}

	// Build markdown table
	table := BuildMarkdownTable(items)

	if table == "" {
		return header
	}

	return header + "\n\n" + table
}

// formatSingle formats a single item as a bullet list with key-value pairs.
// Header format: "{name} ({id}):" if name exists, otherwise "{id}:"
// Fields are sorted alphabetically for consistency.
func formatSingle(item map[string]interface{}) string {
	id := extractID(item)
	name := extractName(item)

	// Build header
	var header string
	if name != "" {
		header = fmt.Sprintf("%s (%s):", name, id)
	} else if id != "" {
		header = fmt.Sprintf("%s:", id)
	} else {
		header = "Resource:"
	}

	// Sort keys alphabetically for consistency
	keys := make([]string, 0, len(item))
	for k := range item {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build bullet list
	var lines []string
	lines = append(lines, header)
	for _, key := range keys {
		value := item[key]
		lines = append(lines, fmt.Sprintf("• %s: %v", key, value))
	}

	return strings.Join(lines, "\n")
}

