package test

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/confluentinc/bincover"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
	testserver "github.com/confluentinc/cli/test/test-server"
)

var (
	noRebuild    = flag.Bool("no-rebuild", false, "skip rebuilding CLI if it already exists")
	update       = flag.Bool("update", false, "update golden files")
	debug        = flag.Bool("debug", true, "enable verbose output")
	cover        = false
	testBin      = "bin/confluent_test"
	covCollector *bincover.CoverageCollector
)

const (
	testBinRace            = "bin/confluent_test_race"
	mergedCoverageFilename = "integ_coverage.txt"
)

// CLITest represents a test configuration
type CLITest struct {
	// Name to show in go test output; defaults to args if not set
	name string
	// The CLI command being tested; this is a string of args and flags passed to the binary
	args string
	// The set of environment variables to be set when the CLI is run
	env []string
	// "default" if you need to login, or "" otherwise
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
	// Optional functions that will be executed directly before the command is run (i.e. overwriting stdin before run)
	preCmdFuncs []bincover.PreCmdFunc
	// Optional functions that will be executed directly after the command is run
	postCmdFuncs []bincover.PostCmdFunc
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

func init() {
	cover = os.Getenv("INTEG_COVER") == "on"

	if os.Getenv("CI") == "on" {
		testBin = testBinRace
	}

	if runtime.GOOS == "windows" {
		testBin = testBin + ".exe"
	}
}

// SetupSuite builds the CLI binary to test
func (s *CLITestSuite) SetupSuite() {
	covCollector = bincover.NewCoverageCollector(mergedCoverageFilename, cover)
	req := require.New(s.T())
	err := covCollector.Setup()
	req.NoError(err)
	s.TestBackend = testserver.StartTestBackend(s.T(), false) // by default do not disable audit-log
	os.Setenv("DISABLE_AUDIT_LOG", "false")
	// dumb but effective
	err = os.Chdir("..")
	req.NoError(err)

	err = os.Setenv("XX_CCLOUD_RBAC_DATAPLANE", "yes")
	req.NoError(err)

	// Temporarily change $HOME, so the current config file isn't altered.
	err = os.Setenv("HOME", os.TempDir())
	req.NoError(err)

	if _, err := os.Stat(binaryPath(s.T(), testBin)); os.IsNotExist(err) || !*noRebuild {
		var makeArgs string
		if testBin == testBinRace {
			makeArgs = "build-integ-race"
		} else {
			makeArgs = "build-integ-nonrace"
		}
		makeCmd := exec.Command("make", makeArgs)
		output, err := makeCmd.CombinedOutput()
		if err != nil {
			s.T().Log(string(output))
			req.NoError(err)
		}
	}
}

func (s *CLITestSuite) TearDownSuite() {
	// Merge coverage profiles.
	_ = covCollector.TearDown()
	s.TestBackend.Close()
}

func (s *CLITestSuite) TestCcloudErrors() {
	args := fmt.Sprintf("login --url %s -vvv", s.TestBackend.GetCloudUrl())

	s.T().Run("invalid user or pass", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=incorrect@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(tt, testBin, env, args, 1)
		require.Contains(tt, output, errors.InvalidLoginErrorMsg)
		require.Contains(tt, output, errors.ComposeSuggestionsMessage(errors.CCloudInvalidLoginSuggestions))
	})

	s.T().Run("suspended organization", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=suspended@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(tt, testBin, env, args, 1)
		require.Contains(tt, output, new(ccloud.SuspendedOrganizationError).Error())
		require.Contains(tt, output, errors.SuspendedOrganizationSuggestions)
	})

	s.T().Run("expired token", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=expired@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(tt, testBin, env, args, 0)
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "expired@user.com", "abc-123", "Confluent"))
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default"))
		output = runCommand(tt, testBin, []string{}, "kafka cluster list", 1)
		require.Contains(tt, output, errors.TokenExpiredMsg)
		require.Contains(tt, output, errors.NotLoggedInErrorMsg)
	})

	s.T().Run("malformed token", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=malformed@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(tt, testBin, env, args, 0)
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "malformed@user.com", "abc-123", "Confluent"))
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default"))

		output = runCommand(s.T(), testBin, []string{}, "kafka cluster list", 1)
		require.Contains(tt, output, errors.CorruptedTokenErrorMsg)
		require.Contains(tt, output, errors.ComposeSuggestionsMessage(errors.CorruptedTokenSuggestions))
	})

	s.T().Run("invalid jwt", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=invalid@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(tt, testBin, env, "logout", 0)
		require.Contains(tt, output, errors.LoggedOutMsg)
		output = runCommand(tt, testBin, env, args, 0)
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "invalid@user.com", "abc-123", "Confluent"))
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default"))

		output = runCommand(s.T(), testBin, []string{}, "kafka cluster list", 1)
		require.Contains(tt, output, errors.CorruptedTokenErrorMsg)
		require.Contains(tt, output, errors.ComposeSuggestionsMessage(errors.CorruptedTokenSuggestions))
	})
}

func (s *CLITestSuite) runCcloudTest(tt CLITest) {
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
			resetConfiguration(t)
		}
		loginURL := s.getLoginURL(true, tt)
		if tt.login == "default" {
			env := []string{pauth.ConfluentCloudEmail + "=fake@user.com", pauth.ConfluentCloudPassword + "=pass1"}
			output := runCommand(t, testBin, env, "login --url "+loginURL, 0)
			if *debug {
				fmt.Println(output)
			}
		}

		if tt.useKafka != "" {
			output := runCommand(t, testBin, []string{}, "kafka cluster use "+tt.useKafka, 0)
			if *debug {
				fmt.Println(output)
			}
		}

		if tt.authKafka != "" {
			output := runCommand(t, testBin, []string{}, "api-key create --resource "+tt.useKafka, 0)
			if *debug {
				fmt.Println(output)
			}
			// HACK: we don't have scriptable output yet so we parse it from the table
			key := strings.TrimSpace(strings.Split(strings.Split(output, "\n")[3], "|")[2])
			output = runCommand(t, testBin, []string{}, fmt.Sprintf("api-key use %s --resource %s", key, tt.useKafka), 0)
			if *debug {
				fmt.Println(output)
			}
		}

		covCollectorOptions := parseCmdFuncsToCoverageCollectorOptions(tt.preCmdFuncs, tt.postCmdFuncs)
		output := runCommand(t, testBin, tt.env, tt.args, tt.wantErrCode, covCollectorOptions...)
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

func (s *CLITestSuite) runConfluentTest(tt CLITest) {
	if tt.name == "" {
		tt.name = tt.args
	}
	if strings.HasPrefix(tt.name, "error") {
		tt.wantErrCode = 1
	}
	s.T().Run(tt.name, func(t *testing.T) {
		if !tt.workflow {
			resetConfiguration(t)
		}

		// Executes login command if test specifies
		loginURL := s.getLoginURL(false, tt)
		if tt.login == "default" {
			env := []string{pauth.ConfluentPlatformUsername + "=fake@user.com", pauth.ConfluentPlatformPassword + "=pass1"}
			output := runCommand(t, testBin, env, "login --url "+loginURL, 0)
			if *debug {
				fmt.Println(output)
			}
		}
		covCollectorOptions := parseCmdFuncsToCoverageCollectorOptions(tt.preCmdFuncs, tt.postCmdFuncs)
		output := runCommand(t, testBin, tt.env, tt.args, tt.wantErrCode, covCollectorOptions...)

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

func runCommand(t *testing.T, binaryName string, env []string, args string, wantErrCode int, coverageCollectorOptions ...bincover.CoverageCollectorOption) string {
	output, exitCode, err := covCollector.RunBinary(binaryPath(t, binaryName), "TestRunMain", env, strings.Split(args, " "), coverageCollectorOptions...)
	if err != nil && wantErrCode == 0 {
		require.Failf(t, "unexpected error",
			"exit %d: %s\n%s", exitCode, args, output)
	}
	require.Equal(t, wantErrCode, exitCode, output)
	return output
}

// Parses pre and post CmdFuncs into CoverageCollectorOptions which can be unsed in covCollector.RunBinary()
func parseCmdFuncsToCoverageCollectorOptions(preCmdFuncs []bincover.PreCmdFunc, postCmdFuncs []bincover.PostCmdFunc) []bincover.CoverageCollectorOption {
	if len(preCmdFuncs) == 0 && len(postCmdFuncs) == 0 {
		return []bincover.CoverageCollectorOption{}
	}
	var options []bincover.CoverageCollectorOption
	return append(options, bincover.PreExec(preCmdFuncs...), bincover.PostExec(postCmdFuncs...))
}

// Used for tests needing to overwrite StdIn for mock input
// returns a cmdFunc struct with the StdinPipe functionality and isPreCmdFunc set to true
// takes an io.Reader with the desired input read into it
func stdinPipeFunc(stdinInput io.Reader) bincover.PreCmdFunc {
	return func(cmd *exec.Cmd) error {
		buf, err := ioutil.ReadAll(stdinInput)
		fmt.Printf("%s", buf)
		if err != nil {
			return err
		}
		if len(buf) == 0 {
			return nil
		}
		writer, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		_, err = writer.Write(buf)
		if err != nil {
			return err
		}
		err = writer.Close()
		if err != nil {
			return err
		}
		return nil
	}
}

func resetConfiguration(t *testing.T) {
	// HACK: delete your current config to isolate tests cases for non-workflow tests...
	// probably don't really want to do this or devs will get mad
	cfg := v1.New(new(config.Params))

	err := cfg.Save()
	require.NoError(t, err)
}

func writeFixture(t *testing.T, fixture string, content string) {
	err := ioutil.WriteFile(FixturePath(t, fixture), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func binaryPath(t *testing.T, binaryName string) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	return path.Join(dir, binaryName)
}
