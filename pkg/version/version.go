package version

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
)

type Version struct {
	Binary    string
	Name      string
	Version   string
	Commit    string
	BuildDate string
	UserAgent string // http
	ClientID  string // kafka
}

const (
	CLIName     = "confluent"
	FullCLIName = "Confluent CLI"
)

func NewVersion(version, commit, buildDate string) *Version {
	dashDelimitedName := strings.ReplaceAll(FullCLIName, " ", "-")

	return &Version{
		Binary:    CLIName,
		Name:      FullCLIName,
		Version:   fmt.Sprintf("v%s", version),
		Commit:    commit,
		BuildDate: buildDate,
		UserAgent: fmt.Sprintf("%s/v%s (https://confluent.io; support@confluent.io)", dashDelimitedName, version),
		ClientID:  fmt.Sprintf("%s_v%s", dashDelimitedName, version),
	}
}

func (v *Version) IsReleased() bool {
	semver, err := version.NewSemver(v.Version)
	if err != nil {
		return false
	}
	return semver.GreaterThan(version.Must(version.NewSemver("v0.0.0"))) && semver.Prerelease() == ""
}

// String returns the version in a standardized format
func (v *Version) String() string {
	return fmt.Sprintf(`%s - %s

Version:     %s
Git Ref:     %s
Build Date:  %s
Go Version:  %s (%s/%s)
Development: %s
`,
		v.Binary,
		v.Name,
		v.Version,
		v.Commit,
		v.BuildDate,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		strconv.FormatBool(!v.IsReleased()))
}
