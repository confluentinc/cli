package flink

import (
	"fmt"
	"strings"
	"unicode"
)

// queryColumnOut is one column of the result schema in the -o json/yaml output.
type queryColumnOut struct {
	Name string `json:"name" yaml:"name"`
	Type string `json:"type" yaml:"type"`
}

// escapeControlChars neutralizes control characters (e.g. raw ANSI escape codes)
// before a field value reaches the terminal. JSON/YAML already escape these via
// their own serializers; this is the equivalent for the plain-table renderer,
// which otherwise lets the terminal execute them (recoloring, moving the cursor).
func escapeControlChars(s string) string {
	if !strings.ContainsFunc(s, needsEscape) {
		return s
	}

	var sb strings.Builder
	for _, r := range s {
		if needsEscape(r) {
			fmt.Fprintf(&sb, "\\x%02x", r)
			continue
		}
		sb.WriteRune(r)
	}
	return sb.String()
}

// needsEscape reports whether r is a control character this renderer neutralizes.
// Tab is left as-is: it's a legitimate data character (e.g. inside a multi-line
// VARCHAR) and carries no cursor/injection risk. Newline and carriage return
// stay escaped despite also being "legitimate" — a raw one in a table cell
// breaks the layout or lets a value inject fake rows — as do ANSI/cursor escapes
// like \x1b, the original reason this exists.
func needsEscape(r rune) bool {
	return unicode.IsControl(r) && r != '\t'
}
