package formatter

// extractID extracts the ID field from an item.
// Returns empty string if not found or not a string.
func extractID(item map[string]interface{}) string {
	if id, ok := item["id"].(string); ok {
		return id
	}
	return ""
}

// extractName extracts a human-readable name from an item.
// Tries fields in priority order: name, display_name, label, title
// Returns empty string if none found or value is not a string.
func extractName(item map[string]interface{}) string {
	// Try fields in priority order per CONTEXT.md
	nameFields := []string{"name", "display_name", "label", "title"}

	for _, field := range nameFields {
		if value, ok := item[field].(string); ok && value != "" {
			return value
		}
	}

	return ""
}
