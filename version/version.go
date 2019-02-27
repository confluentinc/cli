package version

import (
	"fmt"
	"runtime"
	"strings"
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
