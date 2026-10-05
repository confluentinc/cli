package version

import (
	"regexp"

	"github.com/hashicorp/go-version"
)

// Channel is the release stream a binary was built from. It selects the directory the CLI keeps its
// state in, so a production install, a prerelease under evaluation, and a local build never share
// contexts or credentials.
type Channel int

const (
	// Stable is a published GA release, and the only channel that uses the historical path.
	Stable Channel = iota

	// Prerelease is a published release candidate or preview, tagged with a semver prerelease
	// segment such as v5.0.0-rc1.
	Prerelease

	// Dev is anything built outside the release pipeline.
	Dev
)

// StateDirSuffix is appended to ".confluent" to name the channel's state directory. Only Stable
// returns an empty string, which is what keeps existing installs on the path they already use; any
// unrecognized channel falls through to the dev suffix so an unfamiliar build isolates itself rather
// than sharing production state.
func (c Channel) StateDirSuffix() string {
	switch c {
	case Stable:
		return ""
	case Prerelease:
		return "-prerelease"
	default:
		return "-dev"
	}
}

// publishedPrerelease matches the only prerelease labels the release pipeline publishes: rc, alpha,
// beta, or preview, optionally numbered (rc1, rc.2, beta.3). It is anchored, so a goreleaser snapshot
// cut during an RC cycle (5.0.0-rc1-SNAPSHOT-<sha>) or `git describe` output does not match.
var publishedPrerelease = regexp.MustCompile(`^(rc|alpha|beta|preview)\d*(\.\d+)*$`)

// ChannelOf classifies the version string the linker stamps into main.version.
//
// Only an allowlisted prerelease label counts as a published prerelease; any other prerelease
// segment (a `make build` snapshot, -dirty, -cp1, -nightly) is a build we don't recognize, so it
// isolates itself in the dev directory.
//
// Not folded into Version.IsReleased on purpose: it answers a different question, and treats 0.0.1
// as released.
func ChannelOf(s string) Channel {
	semver, err := version.NewSemver(s)
	if err != nil || semver.Segments()[0] == 0 {
		return Dev
	}

	prerelease := semver.Prerelease()
	switch {
	case prerelease == "":
		return Stable
	case publishedPrerelease.MatchString(prerelease):
		return Prerelease
	default:
		return Dev
	}
}

// processChannel is the channel of the running binary. The version is fixed at link time, so there
// is one answer per process, and command construction needs it before any Config exists.
//
// Dev is the default because an unstamped binary did not come from the release pipeline, which also
// keeps a `go test` process aligned with the test binaries it drives.
var processChannel = Dev

// SetProcessChannel records the running binary's channel. Call it once from main, before loading
// configuration or constructing commands.
func SetProcessChannel(channel Channel) {
	processChannel = channel
}

// ProcessChannel reports the running binary's channel.
func ProcessChannel() Channel {
	return processChannel
}
