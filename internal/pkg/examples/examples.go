package examples

import (
	"fmt"
	"strings"

	pversion "github.com/confluentinc/cli/internal/pkg/version"
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
		if strings.HasPrefix(line, pversion.CLIName) {
			line = "$ " + line
		}
		str.WriteString(fmt.Sprintf("  %s\n", line))
	}
	return strings.TrimSuffix(str.String(), "\n")
}
