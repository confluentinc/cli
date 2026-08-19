package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// The channel a build resolves its state directory from is decided in cmd/confluent/main.go, before
// any command runs. Unit tests cover the classifier and the path it produces, but neither can see
// whether main actually calls it: the integration binary carries no version stamp, so "wired
// correctly" and "not wired at all" both come out as the dev channel.
//
// These build stamped binaries and run them, which is the only level at which that wiring is
// observable. A regression here means every existing customer's login moves on upgrade.

func TestChannelState_StampedBuildsUseSeparateDirectories(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
		absent  string
	}{
		{"GA release keeps the historical path", "9.9.9", ".confluent", ".confluent-dev"},
		{"release candidate is isolated", "9.9.9-rc1", ".confluent-prerelease", ".confluent"},
		{"unstamped build is a local build", "", ".confluent-dev", ".confluent"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()

			runStampedCli(t, buildStampedCli(t, test.version), home)

			require.DirExists(t, filepath.Join(home, test.want))
			require.NoDirExists(t, filepath.Join(home, test.absent))
		})
	}
}

// The isolation the feature promises, rather than the paths it happens to pick: state written by
// one channel must be invisible to another sharing the same home directory.
func TestChannelState_ReleaseConfigIsInvisibleToLocalBuild(t *testing.T) {
	home := t.TempDir()
	release := buildStampedCli(t, "9.9.9")
	local := buildStampedCli(t, "")

	runStampedCli(t, release, home, "configuration", "update", "disable_update_check", "true")
	before := readFile(t, filepath.Join(home, ".confluent", "config.json"))
	runStampedCli(t, local, home, "configuration", "update", "disable_update_check", "true")

	require.Equal(t, before, readFile(t, filepath.Join(home, ".confluent", "config.json")),
		"a local build must not modify the release build's configuration")
	require.FileExists(t, filepath.Join(home, ".confluent-dev", "config.json"))
}

// buildStampedCli compiles the CLI with the given main.version, or with none when version is empty.
func buildStampedCli(t *testing.T, version string) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "confluent")
	args := []string{"build", "-o", binary}
	if version != "" {
		args = append(args, "-ldflags=-X main.version="+version)
	}
	args = append(args, "../cmd/confluent")

	output, err := exec.Command("go", args...).CombinedOutput()
	require.NoError(t, err, "go build failed: %s", output)

	return binary
}

func runStampedCli(t *testing.T, binary, home string, args ...string) {
	t.Helper()

	if len(args) == 0 {
		args = []string{"version"}
	}

	cmd := exec.Command(binary, args...)
	// USERPROFILE covers Windows, where os.UserHomeDir ignores HOME.
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s %v failed: %s", binary, args, output)
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()

	contents, err := os.ReadFile(path)
	require.NoError(t, err)

	return contents
}
