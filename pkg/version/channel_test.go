package version

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChannelOf(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    Channel
	}{
		{"GA release", "4.72.0", Stable},
		{"GA release with a v prefix", "v4.72.0", Stable},
		{"release candidate", "5.0.0-rc1", Prerelease},
		{"release candidate with a v prefix", "v5.0.0-rc1", Prerelease},
		{"preview", "5.0.0-preview.2", Prerelease},
		{"beta", "5.0.0-beta3", Prerelease},
		{"unrecognized prerelease marker", "5.0.0-nightly.4", Prerelease},
		{"build metadata on a GA tag", "4.72.0+dirty", Stable},
		// What `make build` actually stamps. It carries a prerelease segment but is a local build,
		// so misreading it as Prerelease would drop developers into the testers' state directory.
		{"goreleaser snapshot", "4.72.0-SNAPSHOT-d962911bb", Dev},
		{"goreleaser snapshot, lowercased", "4.72.0-snapshot-d962911bb", Dev},
		// goreleaser does not strip the tag's prerelease segment, so during an RC cycle a local
		// build carries both. The snapshot marker has to win, or every developer lands in the
		// prerelease directory precisely when real testers are using it.
		{"snapshot built during an RC cycle", "5.0.0-rc1-SNAPSHOT-d962911bb", Dev},
		{"snapshot built during a beta cycle", "5.0.0-beta.1-SNAPSHOT-d962911bb", Dev},
		{"bare go build, nothing stamped", "0.0.0", Dev},
		{"explicit dev stamp", "0.0.0-dev-a1b2c3d", Dev},
		{"any 0.x is a local build", "0.9.1", Dev},
		{"unparsable", "not-a-version", Dev},
		{"empty", "", Dev},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, ChannelOf(test.version))
		})
	}
}

func TestChannel_StateDirSuffix(t *testing.T) {
	req := require.New(t)

	req.Empty(Stable.StateDirSuffix(), "stable must keep using the existing ~/.confluent path")
	req.Equal("-prerelease", Prerelease.StateDirSuffix())
	req.Equal("-dev", Dev.StateDirSuffix())
}

func TestProcessChannel_DefaultsToDev(t *testing.T) {
	require.Equal(t, Dev, ProcessChannel(), "a binary with no version stamped in must not touch production state")
}

func TestSetProcessChannel(t *testing.T) {
	t.Cleanup(func() { SetProcessChannel(Dev) })

	SetProcessChannel(Stable)

	require.Equal(t, Stable, ProcessChannel())
}
