package version

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/terminal"
)

type Version struct {
	Version   string
	Commit    string
	BuildDate string
	BuildHost string
	UserAgent string
}

func NewVersion(version, commit, buildDate, buildHost string) *Version {
	return &Version{
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
		BuildHost: buildHost,
		UserAgent: fmt.Sprintf("Confluent/1.0 ccloud/%s (%s/%s)", version, runtime.GOOS, runtime.GOARCH),
	}
}

func (v *Version) IsReleased() bool {
	return v.Version != "0.0.0" && !strings.Contains(v.Version, "dirty") && !strings.Contains(v.Version, "-g")
}

// PrintVersion prints the version to the prompt in a standardized way
func (v *Version) Print(prompt terminal.Prompt) {
	_, _ = prompt.Printf(`ccloud - Confluent Cloud CLI

Version:     %s
Git Ref:     %s
Build Date:  %s
Build Host:  %s
Go Version:  %s (%s/%s)
Development: %s
`, v.Version,
		v.Commit,
		v.BuildDate,
		v.BuildHost,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		strconv.FormatBool(!v.IsReleased()))
}
