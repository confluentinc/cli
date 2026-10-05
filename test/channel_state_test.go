package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

			s.runStampedCli(s.stampedCli(test.version), home)

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

	s.runStampedCli(release, home, "configuration", "update", "disable_update_check", "true")
	before := s.readFile(filepath.Join(home, ".confluent", "config.json"))
	s.runStampedCli(local, home, "configuration", "update", "disable_update_check", "true")

	s.Require().Equal(before, s.readFile(filepath.Join(home, ".confluent", "config.json")),
		"a local build must not modify the release build's configuration")
	s.Require().FileExists(filepath.Join(home, ".confluent-dev", "config.json"))
}

// stampedClis caches binaries by version for the whole suite, so each one is built once.
var stampedClis = map[string]string{}

// stampedCli returns the CLI compiled with the given main.version, or with none when version is
// empty.
func (s *CLITestSuite) stampedCli(version string) string {
	if binary, ok := stampedClis[version]; ok {
		return binary
	}

	// rootT outlives this method, so the cached binary is still there for the next one.
	binary := filepath.Join(s.rootT.TempDir(), "confluent")
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
	output, err := exec.Command("go", "build", "-ldflags="+ldflags, "-o", binary, "./cmd/confluent").CombinedOutput()
	s.Require().NoError(err, "go build failed: %s", output)

	stampedClis[version] = binary
	return binary
}

func (s *CLITestSuite) runStampedCli(binary, home string, args ...string) {
	if len(args) == 0 {
		args = []string{"version"}
	}

	cmd := exec.Command(binary, args...)
	// USERPROFILE covers Windows, where os.UserHomeDir ignores HOME.
	cmd.Env = append(os.Environ(), "HOME="+home, "USERPROFILE="+home)

	output, err := cmd.CombinedOutput()
	s.Require().NoError(err, "%s %v failed: %s", binary, args, output)
}

func (s *CLITestSuite) readFile(path string) []byte {
	contents, err := os.ReadFile(path)
	s.Require().NoError(err)

	return contents
}
