package test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/google/shlex"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pauth "github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/utils"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

var (
	update  = flag.Bool("update", false, "update golden files")
	debug   = flag.Bool("debug", true, "enable verbose output")
	testBin = "test/bin/confluent"
)

// CLITest represents a test configuration
type CLITest struct {
	// Name to show in go test output; defaults to args if not set
	name string
	// The CLI command being tested; this is a string of args and flags passed to the binary
	args string
	// The set of environment variables to be set when the CLI is run
	env []string
	// The login context; either "cloud" or "onprem"
	login string
	// Optional Cloud URL if test does not use default server
	loginURL string
	// The kafka cluster ID to "use"
	useKafka string
	// Create and use an API Key to set as Kafka credentials
	authKafka bool
	// Name of a golden output fixture containing expected output
	fixture string
	// True if audit-log is disabled
	disableAuditLog bool
	// True iff fixture represents a regex
	regex bool
	// True iff testing plugins
	arePluginsEnabled bool
	// Fixed string to check if output contains
	contains string
	// Fixed string to check that output does not contain
	notContains string
	// Expected exit code (e.g., 0 for success or 1 for failure)
	exitCode int
	// If true, don't reset the config/state between tests to enable testing CLI workflows
	workflow bool
	// If true, run the CLI from a private copy of the test binary; for tests which replace
	// the running binary, such as `confluent update`
	isolatedBin bool
	// An optional function that allows you to specify other calls
	wantFunc func(t *testing.T)
	input    string
}

// CLITestSuite is the CLI integration tests.
type CLITestSuite struct {
	suite.Suite
	TestBackend *testserver.TestBackend
}

// TestCLI runs the CLI integration test suite.
func TestCLI(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	req := require.New(s.T())

	// dumb but effective
	err := os.Chdir("..")
	req.NoError(err)

	target := "build-for-integration-test"
	if runtime.GOOS == "windows" {
		target += "-windows"
		testBin += ".exe"
	}

	output, err := exec.Command("make", target).CombinedOutput()
	req.NoError(err, string(output))

	s.TestBackend = testserver.StartTestBackend(s.T(), true) // by default do not disable audit-log
	os.Setenv("DISABLE_AUDIT_LOG", "false")

	config.SetTempHomeDir()
}

func (s *CLITestSuite) TearDownSuite() {
	s.TestBackend.Close()
}

func (s *CLITestSuite) runIntegrationTest(test CLITest) {
	if test.name == "" {
		test.name = test.args
	}

	s.T().Run(test.name, func(t *testing.T) {
		isAuditLogDisabled := os.Getenv("DISABLE_AUDIT_LOG") == "true"
		if isAuditLogDisabled != test.disableAuditLog {
			s.TestBackend.Close()
			os.Setenv("DISABLE_AUDIT_LOG", strconv.FormatBool(test.disableAuditLog))
			s.TestBackend = testserver.StartTestBackend(t, !test.disableAuditLog)
		}

		if !test.workflow {
			resetConfiguration(t, test.arePluginsEnabled)
		}

		bin := testBin
		if test.isolatedBin {
			bin = copyTestBin(t)
		}

		// Executes login command if test specifies
		switch test.login {
		case "cloud":
			loginString := fmt.Sprintf("login --url %s", s.getLoginURL(true, test))
			env := append([]string{pauth.ConfluentCloudEmail + "=fake@user.com", pauth.ConfluentCloudPassword + "=pass1"}, test.env...)
			for _, e := range env {
				keyVal := strings.Split(e, "=")
				os.Setenv(keyVal[0], keyVal[1])
			}

			defer func() {
				for _, e := range env {
					keyVal := strings.Split(e, "=")
					os.Unsetenv(keyVal[0])
				}
			}()

			output := runCommand(t, bin, env, loginString, 0, "")
			if *debug {
				fmt.Println(output)
			}
		case "onprem":
			loginURL := s.getLoginURL(false, test)
			env := []string{pauth.ConfluentPlatformUsername + "=fake@user.com", pauth.ConfluentPlatformPassword + "=pass1"}
			output := runCommand(t, bin, env, "login --url "+loginURL, 0, "")
			if *debug {
				fmt.Println(output)
			}
		}

		if test.useKafka != "" {
			output := runCommand(t, bin, []string{}, fmt.Sprintf("kafka cluster use %s", test.useKafka), 0, "")
			if *debug {
				fmt.Println(output)
			}
		}

		if test.authKafka {
			output := runCommand(t, bin, []string{}, fmt.Sprintf("api-key create --resource %s --use", test.useKafka), 0, "")
			if *debug {
				fmt.Println(output)
			}
		}

		output := runCommand(t, bin, test.env, test.args, test.exitCode, test.input)
		if *debug {
			fmt.Println(output)
		}

		s.validateTestOutput(test, t, output)
	})
}

func (s *CLITestSuite) getLoginURL(isCloud bool, test CLITest) string {
	if test.loginURL != "" {
		return test.loginURL
	}

	if isCloud {
		return s.TestBackend.GetCloudUrl()
	} else {
		return s.TestBackend.GetMdsUrl()
	}
}

func (s *CLITestSuite) validateTestOutput(test CLITest, t *testing.T, output string) {
	if *update && !test.regex && test.fixture != "" {
		writeFixture(t, test.fixture, output)
	}
	actual := utils.NormalizeNewLines(output)
	if test.contains != "" {
		require.Contains(t, actual, test.contains)
	} else if test.notContains != "" {
		require.NotContains(t, actual, test.notContains)
	} else if test.fixture != "" {
		expected := utils.NormalizeNewLines(LoadFixture(t, test.fixture))
		if test.regex {
			require.Regexp(t, expected, actual)
		} else {
			require.Equal(t, expected, actual)
		}
	}
	if test.wantFunc != nil {
		test.wantFunc(t)
	}
}

// copyTestBin copies the test binary into its own temporary directory and returns the
// absolute path to the copy. `confluent update` replaces the running binary and leaves a
// `.confluent.exe.old` sidecar beside it; on Windows that sidecar stays locked by the
// process which just exited, so consecutive updates sharing a directory fail to rename
// over it. Giving each test its own directory keeps those sidecars from colliding, and
// leaves the shared test binary untouched.
func copyTestBin(t *testing.T) string {
	// SetupSuite has already suffixed testBin with ".exe" on Windows.
	binary, err := os.ReadFile(testBin)
	require.NoError(t, err)

	dir, err := os.MkdirTemp("", "confluent-test-bin")
	require.NoError(t, err)

	// Best effort: on Windows the sidecar may still be locked by the process which just
	// exited, and failing to clean up a temporary directory shouldn't fail the test.
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	path := filepath.Join(dir, filepath.Base(testBin))
	require.NoError(t, os.WriteFile(path, binary, 0755))

	return path
}

func runCommand(t *testing.T, binaryName string, env []string, argString string, exitCode int, input string) string {
	dir, err := os.Getwd()
	require.NoError(t, err)

	// HACK: google/shlex does not support non-POSIX shell parsing
	if runtime.GOOS == "windows" {
		argString = strings.ReplaceAll(argString, `\'`, "SINGLE QUOTE")
		argString = strings.ReplaceAll(argString, `\"`, "DOUBLE QUOTE")
		argString = strings.ReplaceAll(argString, `\`, `\\`)
		argString = strings.ReplaceAll(argString, "SINGLE QUOTE", `\'`)
		argString = strings.ReplaceAll(argString, "DOUBLE QUOTE", `\"`)
	}

	args, err := shlex.Split(argString)
	require.NoError(t, err)

	binaryPath := binaryName
	if !filepath.IsAbs(binaryPath) {
		binaryPath = filepath.Join(dir, binaryPath)
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = strings.NewReader(input)

	out, err := cmd.CombinedOutput()
	if exitCode == 0 {
		require.NoError(t, err, string(out))
	}
	require.Equal(t, exitCode, cmd.ProcessState.ExitCode(), string(out))

	return string(out)
}

func resetConfiguration(t *testing.T, arePluginsEnabled bool) {
	// HACK: delete your current config to isolate tests cases for non-workflow tests...
	// probably don't really want to do this or devs will get mad
	cfg := config.New()
	cfg.DisablePlugins = !arePluginsEnabled
	cfg.EnableColor = false
	err := cfg.Save()
	require.NoError(t, err)
}

func writeFixture(t *testing.T, fixture, content string) {
	path := fixturePath(t, fixture)

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
