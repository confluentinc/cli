package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes ANSI color codes",
			input:    "\x1b[31mError\x1b[0m",
			expected: "Error",
		},
		{
			name:     "removes multiple color codes",
			input:    "\x1b[32mSuccess:\x1b[0m \x1b[1mBold\x1b[0m",
			expected: "Success: Bold",
		},
		{
			name:     "removes cursor positioning codes",
			input:    "\x1b[2J\x1b[HCleared screen",
			expected: "Cleared screen",
		},
		{
			name:     "plain text unchanged",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "removes SGR codes (Select Graphic Rendition)",
			input:    "\x1b[1;31mBold Red\x1b[0m",
			expected: "Bold Red",
		},
		{
			name:     "handles complex ANSI sequences",
			input:    "\x1b[2K\x1b[1GText with clear line",
			expected: "Text with clear line",
		},
		{
			name:     "complex lipgloss output",
			input:    "\x1b[38;2;255;100;50m\x1b[1mStyled\x1b[0m",
			expected: "Styled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripANSI(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
