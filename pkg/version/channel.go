package version

import (
	"strings"

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

func (c Channel) String() string {
	switch c {
	case Stable:
		return "stable"
	case Prerelease:
		return "prerelease"
	default:
		return "dev"
	}
}

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

// snapshotMarker is what goreleaser puts in a local build's version. Its default template is
// "{{ .Version }}-SNAPSHOT-{{ .ShortCommit }}" and it does not strip an existing prerelease segment,
// so during a release-candidate cycle `make build` stamps 5.0.0-rc1-SNAPSHOT-<sha>.
const snapshotMarker = "SNAPSHOT"

// ChannelOf classifies the version string the linker stamps into main.version.
//
// A prerelease segment alone cannot mean "published prerelease", since `make build` puts one on
// every developer's binary; the snapshot marker is what separates them. Anything else carrying one
// is treated as published, which errs toward isolating an unfamiliar build. That includes a
// Confluent Platform suffix like 4.72.0-cp1, which no tag in this repo's history uses.
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
	case hasSnapshotToken(prerelease):
		return Dev
	default:
		return Prerelease
	}
}

// hasSnapshotToken reports whether the snapshot marker appears as a whole segment of the prerelease,
// not merely as a substring, so a published label like 5.0.0-presnapshot stays a prerelease. The
// segment is the goreleaser template's own `-SNAPSHOT-<sha>`, joined with hyphens onto any existing
// prerelease dot-identifiers, so both delimiters bound a token.
func hasSnapshotToken(prerelease string) bool {
	for _, token := range strings.FieldsFunc(prerelease, func(r rune) bool { return r == '-' || r == '.' }) {
		if strings.EqualFold(token, snapshotMarker) {
			return true
		}
	}
	return false
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
