package shared

import (
	"runtime"
	"strconv"

	"github.com/confluentinc/cli/command"
	"github.com/confluentinc/cli/version"
)

// PrintVersion prints the version to the prompt in a standardized way
func PrintVersion(version *version.Version, prompt command.Prompt) {
	_, _ = prompt.Printf(`ccloud - Confluent Cloud CLI

Version:     %s
Git Ref:     %s
Build Date:  %s
Build Host:  %s
Go Version:  %s (%s/%s)
Development: %s
`, version.Version,
		version.Commit,
		version.BuildDate,
		version.BuildHost,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		strconv.FormatBool(!version.IsReleased()))
}
