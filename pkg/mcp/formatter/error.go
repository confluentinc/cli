package formatter

import (
	"strings"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// FormatError creates a human-readable error message with command context and suggestions.
// It follows these rules:
// - If rawOutput starts with "Error:", return it as-is (already formatted)
// - Otherwise, build error message with optional command context
// - Include suggestions if error implements ErrorWithSuggestions
// - Use text-only formatting (no symbols/emoji)
func FormatError(err error, commandPath string, rawOutput string) string {
	var builder strings.Builder

	// If rawOutput already starts with "Error:", pass through
	if strings.HasPrefix(rawOutput, "Error:") {
		return rawOutput
	}

	// Add command context if provided
	if commandPath != "" {
		builder.WriteString("Command: ")
		builder.WriteString(commandPath)
		builder.WriteString("\n")
	}

	// Add error message
	builder.WriteString("Error: ")
	builder.WriteString(err.Error())
	builder.WriteString("\n")

	// Add suggestions if available
	suggestions := errors.DisplaySuggestionsMessage(err)
	if suggestions != "" {
		builder.WriteString(suggestions)
	}

	return builder.String()
}
