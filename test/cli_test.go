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
	"strings"
	"testing"

	"github.com/confluentinc/bincover"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/config"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
	testserver "github.com/confluentinc/cli/test/test-server"
)

var (
	noRebuild           = flag.Bool("no-rebuild", false, "skip rebuilding CLI if it already exists")
	update              = flag.Bool("update", false, "update golden files")
	debug               = flag.Bool("debug", true, "enable verbose output")
	skipSsoBrowserTests = flag.Bool("skip-sso-browser-tests", false, "If flag is preset, run the tests that require a web browser.")
	ssoTestEmail        = *flag.String("sso-test-user-email", "ziru+paas-integ-sso@confluent.io", "The email of an sso enabled test user.")
	ssoTestPassword     = *flag.String("sso-test-user-password", "aWLw9eG+F", "The password for the sso enabled test user.")
	// this connection is preconfigured in Auth0 to hit a test Okta account
	ssoTestConnectionName = *flag.String("sso-test-connection-name", "confluent-dev", "The Auth0 SSO connection name.")
	// browser tests by default against devel
	ssoTestLoginUrl  = *flag.String("sso-test-login-url", "https://devel.cpdev.cloud", "The login url to use for the sso browser test.")
	cover            = false
	ccloudTestBin    = ccloudTestBinNormal
	confluentTestBin = confluentTestBinNormal
	covCollector     *bincover.CoverageCollector
)

const (
	confluentTestBinNormal = "confluent_test"
	ccloudTestBinNormal    = "ccloud_test"
	ccloudTestBinRace      = "ccloud_test_race"
	confluentTestBinRace   = "confluent_test_race"
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
	collectCoverage := os.Getenv("INTEG_COVER")
	cover = collectCoverage == "on"
	ciEnv := os.Getenv("CI")
	if ciEnv == "on" {
		ccloudTestBin = ccloudTestBinRace
		confluentTestBin = confluentTestBinRace
	}
	if runtime.GOOS == "windows" {
		ccloudTestBin = ccloudTestBin + ".exe"
		confluentTestBin = confluentTestBin + ".exe"
	}
}

// SetupSuite builds the CLI binary to test
func (s *CLITestSuite) SetupSuite() {
	covCollector = bincover.NewCoverageCollector(mergedCoverageFilename, cover)
	req := require.New(s.T())
	err := covCollector.Setup()
	req.NoError(err)
	s.TestBackend = testserver.StartTestBackend(s.T())

	// dumb but effective
	err = os.Chdir("..")
	req.NoError(err)
	for _, binary := range []string{ccloudTestBin, confluentTestBin} {
		if _, err = os.Stat(binaryPath(s.T(), binary)); os.IsNotExist(err) || !*noRebuild {
			var makeArgs string
			if ccloudTestBin == ccloudTestBinRace {
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
}

func (s *CLITestSuite) TearDownSuite() {
	// Merge coverage profiles.
	_ = covCollector.TearDown()
	s.TestBackend.Close()
}

func (s *CLITestSuite) TestCcloudErrors() {
	// TODO: Add this test back when we add prompt testing for integration test
	// Now that non-interactive login is officially supported, we ignore failures from env var and netrc login and give user another chance at login
	//	s.T().Run("invalid user or pass", func(tt *testing.T) {
	//		loginURL := serveErrors(tt)
	//		env := []string{fmt.Sprintf("%s=incorrect@user.com", pauth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", pauth.CCloudPasswordEnvVar)}
	//		output := runCommand(tt, ccloudTestBin, env, "login --url "+loginURL, 1)
	//		require.Contains(tt, output, errors.InvalidLoginErrorMsg)
	//		require.Contains(tt, output, errors.ComposeSuggestionsMessage(errors.CCloudInvalidLoginSuggestions))
	//	})

	s.T().Run("expired token", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=expired@user.com", pauth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", pauth.CCloudPasswordEnvVar)}
		output := runCommand(tt, ccloudTestBin, env, "login -vvv --url "+s.TestBackend.GetCloudUrl(), 0)
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInAsMsg, "expired@user.com"))
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default"))
		output = runCommand(tt, ccloudTestBin, []string{}, "kafka cluster list", 1)
		require.Contains(tt, output, errors.TokenExpiredMsg)
		require.Contains(tt, output, errors.NotLoggedInErrorMsg)
	})

	s.T().Run("malformed token", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=malformed@user.com", pauth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", pauth.CCloudPasswordEnvVar)}
		output := runCommand(tt, ccloudTestBin, env, "login -vvv --url "+s.TestBackend.GetCloudUrl(), 0)
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInAsMsg, "malformed@user.com"))
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default"))

		output = runCommand(s.T(), ccloudTestBin, []string{}, "kafka cluster list", 1)
		require.Contains(tt, output, errors.CorruptedTokenErrorMsg)
		require.Contains(tt, output, errors.ComposeSuggestionsMessage(errors.CorruptedTokenSuggestions))
	})

	s.T().Run("invalid jwt", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=invalid@user.com", pauth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", pauth.CCloudPasswordEnvVar)}
		output := runCommand(tt, ccloudTestBin, env, "login -vvv --url "+s.TestBackend.GetCloudUrl(), 0)
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInAsMsg, "invalid@user.com"))
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default"))

		output = runCommand(s.T(), ccloudTestBin, []string{}, "kafka cluster list", 1)
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
		if !tt.workflow {
			resetConfiguration(t, "ccloud")
		}
		loginURL := s.getLoginURL("ccloud", tt)
		if tt.login == "default" {
			env := []string{fmt.Sprintf("%s=fake@user.com", pauth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", pauth.CCloudPasswordEnvVar)}
			output := runCommand(t, ccloudTestBin, env, "login --url "+loginURL, 0)
			if *debug {
				fmt.Println(output)
			}
		}

		if tt.useKafka != "" {
			output := runCommand(t, ccloudTestBin, []string{}, "kafka cluster use "+tt.useKafka, 0)
			if *debug {
				fmt.Println(output)
			}
		}

		if tt.authKafka != "" {
			output := runCommand(t, ccloudTestBin, []string{}, "api-key create --resource "+tt.useKafka, 0)
			if *debug {
				fmt.Println(output)
			}
			// HACK: we don't have scriptable output yet so we parse it from the table
			key := strings.TrimSpace(strings.Split(strings.Split(output, "\n")[3], "|")[2])
			output = runCommand(t, ccloudTestBin, []string{}, fmt.Sprintf("api-key use %s --resource %s", key, tt.useKafka), 0)
			if *debug {
				fmt.Println(output)
			}
		}
		covCollectorOptions := parseCmdFuncsToCoverageCollectorOptions(tt.preCmdFuncs, tt.postCmdFuncs)
		output := runCommand(t, ccloudTestBin, tt.env, tt.args, tt.wantErrCode, covCollectorOptions...)
		if *debug {
			fmt.Println(output)
		}

		if strings.HasPrefix(tt.args, "kafka cluster create") ||
			strings.HasPrefix(tt.args, "config context current") {
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
			resetConfiguration(t, "confluent")
		}

		// Executes login command if test specifies
		loginURL := s.getLoginURL("confluent", tt)
		if tt.login == "default" {
			env := []string{"XX_CONFLUENT_USERNAME=fake@user.com", "XX_CONFLUENT_PASSWORD=pass1"}
			output := runCommand(t, confluentTestBin, env, "login --url "+loginURL, 0)
			if *debug {
				fmt.Println(output)
			}
		}
		covCollectorOptions := parseCmdFuncsToCoverageCollectorOptions(tt.preCmdFuncs, tt.postCmdFuncs)
		output := runCommand(t, confluentTestBin, tt.env, tt.args, tt.wantErrCode, covCollectorOptions...)

		if strings.HasPrefix(tt.args, "config context list") ||
			strings.HasPrefix(tt.args, "config context current") {
			re := regexp.MustCompile("https?://127.0.0.1:[0-9]+")
			output = re.ReplaceAllString(output, "http://127.0.0.1:12345")
		}

		s.validateTestOutput(tt, t, output)
	})
}

func (s *CLITestSuite) getLoginURL(cliName string, tt CLITest) string {
	if tt.loginURL != "" {
		return tt.loginURL
	}
	switch cliName {
	case "ccloud":
		return s.TestBackend.GetCloudUrl()
	case "confluent":
		return s.TestBackend.GetMdsUrl()
	default:
		return ""
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

func resetConfiguration(t *testing.T, cliName string) {
	// HACK: delete your current config to isolate tests cases for non-workflow tests...
	// probably don't really want to do this or devs will get mad
	cfg := v3.New(&config.Params{
		CLIName: cliName,
	})
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
