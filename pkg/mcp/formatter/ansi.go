package formatter

import (
	"github.com/acarl005/stripansi"
)

// StripANSI removes all ANSI escape codes from the input string.
// This includes color codes (SGR), cursor positioning (CSI), and
// operating system commands (OSC) that are commonly used in terminal output.
func StripANSI(s string) string {
	return stripansi.Strip(s)
}
