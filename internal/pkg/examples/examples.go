package examples

import (
	"fmt"
	"strings"
)

type Example struct {
	Text string
	Code string
}

func BuildExampleString(examples ...Example) string {
	var str strings.Builder
	for _, e := range examples {
		str.WriteString(e.Text + "\n\n")
		if e.Code != "" {
			str.WriteString(formatCodeBlock(e.Code) + "\n\n")
		}
	}
	return strings.TrimSuffix(str.String(), "\n\n")
}

// formatCodeBlock prefixes each command with a "$" to represent a shell prompt
func formatCodeBlock(lines string) string {
	var str strings.Builder
	for _, line := range strings.Split(lines, "\n") {
		if strings.HasPrefix(line, "confluent") || strings.HasPrefix(line, "ccloud") {
			line = "$ " + line
		}
		str.WriteString(fmt.Sprintf("  %s\n", line))
	}
	return strings.TrimSuffix(str.String(), "\n")
}
