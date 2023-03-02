package test

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/utils"
	testserver "github.com/confluentinc/cli/test/test-server"
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
	// The login context; either "cloud" or "platform"
	login string
	// Optional Cloud URL if test does not use default server
	loginURL string
	// The kafka cluster ID to "use"
	useKafka string
	// The API Key to set as Kafka credentials
	authKafka string
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
	wantErrCode int
	// If true, don't reset the config/state between tests to enable testing CLI workflows
	workflow bool
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

	if runtime.GOOS == "windows" {
		testBin += ".exe"
	}

	s.TestBackend = testserver.StartTestBackend(s.T(), false) // by default do not disable audit-log
	os.Setenv("DISABLE_AUDIT_LOG", "false")

	// dumb but effective
	err := os.Chdir("..")
	req.NoError(err)

	// Temporarily change $HOME, so the current config file isn't altered.
	err = os.Setenv("HOME", os.TempDir())
	req.NoError(err)
}

func (s *CLITestSuite) TearDownSuite() {
	s.TestBackend.Close()
}

func (s *CLITestSuite) runIntegrationTest(tt CLITest) {
	if tt.name == "" {
		tt.name = tt.args
	}

	if strings.HasPrefix(tt.name, "error") {
		tt.wantErrCode = 1
	}

	s.T().Run(tt.name, func(t *testing.T) {
		isAuditLogDisabled := os.Getenv("DISABLE_AUDIT_LOG") == "true"
		if isAuditLogDisabled != tt.disableAuditLog {
			s.TestBackend.Close()
			s.TestBackend = nil
			os.Setenv("DISABLE_AUDIT_LOG", strconv.FormatBool(tt.disableAuditLog))
			s.TestBackend = testserver.StartTestBackend(t, tt.disableAuditLog)
		}

		if !tt.workflow {
			resetConfiguration(t, tt.arePluginsEnabled)
		}

		// Executes login command if test specifies
		switch tt.login {
		case "cloud":
			loginString := fmt.Sprintf("login --url %s", s.getLoginURL(true, tt))
			env := append([]string{pauth.ConfluentCloudEmail + "=fake@user.com", pauth.ConfluentCloudPassword + "=pass1"}, tt.env...)
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

			output := runCommand(t, testBin, env, loginString, 0, "")
			if *debug {
				fmt.Println(output)
			}
		case "platform":
			loginURL := s.getLoginURL(false, tt)
			env := []string{pauth.ConfluentPlatformUsername + "=fake@user.com", pauth.ConfluentPlatformPassword + "=pass1"}
			output := runCommand(t, testBin, env, "login --url "+loginURL, 0, "")
			if *debug {
				fmt.Println(output)
			}
		}

		if tt.useKafka != "" {
			output := runCommand(t, testBin, []string{}, "kafka cluster use "+tt.useKafka, 0, "")
			if *debug {
				fmt.Println(output)
			}
		}

		if tt.authKafka != "" {
			output := runCommand(t, testBin, []string{}, fmt.Sprintf("api-key create --resource %s --use", tt.useKafka), 0, "")
			if *debug {
				fmt.Println(output)
			}
		}

		output := runCommand(t, testBin, tt.env, tt.args, tt.wantErrCode, tt.input)
		if *debug {
			fmt.Println(output)
		}

		if strings.HasPrefix(tt.args, "kafka cluster create") {
			re := regexp.MustCompile("https?://127.0.0.1:[0-9]+")
			output = re.ReplaceAllString(output, "http://127.0.0.1:12345")
		}

		s.validateTestOutput(tt, t, output)
	})
}

func (s *CLITestSuite) getLoginURL(isCloud bool, tt CLITest) string {
	if tt.loginURL != "" {
		return tt.loginURL
	}

	if isCloud {
		return s.TestBackend.GetCloudUrl()
	} else {
		return s.TestBackend.GetMdsUrl()
	}
}

func (s *CLITestSuite) validateTestOutput(tt CLITest, t *testing.T, output string) {
	if *update && !tt.regex && tt.fixture != "" {
		writeFixture(t, tt.fixture, output)
	}
	actual := utils.NormalizeNewLines(output)
	if tt.contains != "" {
		require.Contains(t, actual, tt.contains)
	} else if tt.notContains != "" {
		require.NotContains(t, actual, tt.notContains)
	} else if tt.fixture != "" {
		expected := utils.NormalizeNewLines(LoadFixture(t, tt.fixture))
		if tt.regex {
			require.Regexp(t, expected, actual)
		} else if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("\n   actual:\n%s\nexpected:\n%s", actual, expected)
		}
	}
	if tt.wantFunc != nil {
		tt.wantFunc(t)
	}
}

func runCommand(t *testing.T, binaryName string, env []string, args string, wantErrCode int, input string) string {
	dir, err := os.Getwd()
	require.NoError(t, err)

	cmd := exec.Command(filepath.Join(dir, binaryName), strings.Split(args, " ")...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = strings.NewReader(input)

	out, err := cmd.CombinedOutput()
	require.Equal(t, wantErrCode, cmd.ProcessState.ExitCode())
	if wantErrCode == 0 {
		require.NoError(t, err)
	}

	return string(out)
}

func resetConfiguration(t *testing.T, arePluginsEnabled bool) {
	// HACK: delete your current config to isolate tests cases for non-workflow tests...
	// probably don't really want to do this or devs will get mad
	cfg := v1.New()
	cfg.DisablePlugins = !arePluginsEnabled
	err := cfg.Save()
	require.NoError(t, err)
}

func writeFixture(t *testing.T, fixture string, content string) {
	err := os.WriteFile(FixturePath(t, fixture), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func unsetFreeTrialEnv() {
	os.Unsetenv("IS_ON_FREE_TRIAL")
}

func unsetMarketplaceOrgEnv() {
	os.Unsetenv("IS_ORG_ON_MARKETPLACE")
}
