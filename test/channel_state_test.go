package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
//
// Each run gets its own home rather than the suite's, which SetupTest seeds with a dev-channel
// config.json that would satisfy or break these directory assertions on its own.

func (s *CLITestSuite) TestChannelState_StampedBuildsUseSeparateDirectories() {
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
		s.Run(test.name, func() {
			home := s.T().TempDir()

			runStampedCli(s.T(), s.stampedCli(test.version), home)

			s.Require().DirExists(filepath.Join(home, test.want))
			s.Require().NoDirExists(filepath.Join(home, test.absent))
		})
	}
}

// The isolation the feature promises, rather than the paths it happens to pick: state written by
// one channel must be invisible to another sharing the same home directory.
func (s *CLITestSuite) TestChannelState_ReleaseConfigIsInvisibleToLocalBuild() {
	home := s.T().TempDir()
	release := s.stampedCli("9.9.9")
	local := s.stampedCli("")

	stores := []string{"settings.json", "contexts.json"}
	runStampedCli(s.T(), release, home, "configuration", "update", "disable_update_check", "true")
	before := map[string][]byte{}
	for _, store := range stores {
		before[store] = readFile(s.T(), filepath.Join(home, ".confluent", store))
	}
	runStampedCli(s.T(), local, home, "configuration", "update", "disable_update_check", "true")

	for _, store := range stores {
		s.Require().Equal(before[store], readFile(s.T(), filepath.Join(home, ".confluent", store)),
			"a local build must not modify the release build's %s", store)
		s.Require().FileExists(filepath.Join(home, ".confluent-dev", store))
	}
}

// stampedCli returns the CLI compiled with the given main.version, or with none when version is
// empty.
func (s *CLITestSuite) stampedCli(version string) string {
	binary := filepath.Join(s.T().TempDir(), "confluent")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	// isTest keeps the stamped binary off the real update service, which `version` and
	// `configuration update` would otherwise hit; the normal integration build stamps it too.
	ldflags := "-X main.isTest=true"
	if version != "" {
		ldflags += " -X main.version=" + version
	}

	// SetupSuite has already changed to the repo root.
	cmd := exec.Command("go", "build", "-ldflags="+ldflags, "-o", binary, "./cmd/confluent")
	cmd.Env = append(os.Environ(), s.goEnv...)
	output, err := cmd.CombinedOutput()
	s.Require().NoError(err, "go build failed: %s", output)

	return binary
}

// runStampedCli runs binary against home (`version` when no args are given) and returns its output.
func runStampedCli(t *testing.T, binary, home string, args ...string) string {
	t.Helper()
	if len(args) == 0 {
		args = []string{"version"}
	}

	cmd := exec.Command(binary, args...)
	// USERPROFILE covers Windows, where os.UserHomeDir ignores HOME.
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)

	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "%s %v failed: %s", binary, args, output)

	return string(output)
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	contents, err := os.ReadFile(path)
	require.NoError(t, err)

	return contents
}
