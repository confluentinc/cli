package test

import (
	"flag"
	"fmt"
	"github.com/confluentinc/bincover"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/config"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/utils"
	test_server "github.com/confluentinc/cli/test/test-server"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var (
	noRebuild           = flag.Bool("no-rebuild", false, "skip rebuilding CLI if it already exists")
	update              = flag.Bool("update", false, "update golden files")
	debug               = true //= flag.Bool("debug", true, "enable verbose output")
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
	Name string
	// The CLI command being tested; this is a string of args and flags passed to the binary
	Args string
	// The set of environment variables to be set when the CLI is run
	Env []string
	// "default" if you need to login, or "" otherwise
	Login string
	// Optional Cloud URL if test does not use default server
	loginURL string
	// The kafka cluster ID to "use"
	UseKafka string
	// The API Key to set as Kafka credentials
	AuthKafka string
	// Name of a golden output fixture containing expected output
	Fixture string
	// True iff fixture represents a regex
	Regex bool
	// Fixed string to check if output contains
	Contains string
	// Fixed string to check that output does not contain
	NotContains string
	// Expected exit code (e.g., 0 for success or 1 for failure)
	WantErrCode int
	// If true, don't reset the config/state between tests to enable testing CLI workflows
	Workflow bool
	// An optional function that allows you to specify other calls
	WantFunc func(t *testing.T)
	// Optional functions that will be executed directly before the command is run (i.e. overwriting stdin before run)
	PreCmdFuncs []bincover.PreCmdFunc
	// Optional functions that will be executed directly after the command is run
	PostCmdFuncs []bincover.PostCmdFunc
}

// CLITestSuite is the CLI integration tests.
type CLITestSuite struct {
	suite.Suite
	TestBackend *test_server.TestBackend
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

func (s *CLITestSuite) SetupSuite() {
	covCollector = bincover.NewCoverageCollector(mergedCoverageFilename, cover)
	req := require.New(s.T())
	err := covCollector.Setup()
	req.NoError(err)
	s.TestBackend = test_server.StartTestBackend(s.T())

	// dumb but effective
	dir, err := os.Getwd()
	s.T().Log("1: " + dir)
	err = os.Chdir(strings.Split(dir, "/test")[0])
	dir2, err := os.Getwd()
	s.T().Log("2: " + dir2)
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
	//s.TestBackend.Close() // TODO hopefully add this back at some point
}

func (s *CLITestSuite) RunCcloudTest(tt CLITest) {
	if tt.Name == "" {
		tt.Name = tt.Args
	}
	if strings.HasPrefix(tt.Name, "error") {
		tt.WantErrCode = 1
	}

	s.T().Run(tt.Name, func(t *testing.T) {
		if !tt.Workflow {
			//t.Parallel()
			ResetConfiguration(t, "ccloud")
		}
		loginURL := s.getLoginURL("ccloud", tt)
		if tt.Login == "default" {
			env := []string{fmt.Sprintf("%s=fake@user.com", pauth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", pauth.CCloudPasswordEnvVar)}
			output := runCommand(t, ccloudTestBin, env, "login --url "+loginURL, 0)
			if debug {
				fmt.Println(output)
			}
		}

		if tt.UseKafka != "" {
			output := runCommand(t, ccloudTestBin, []string{}, "kafka cluster use "+tt.UseKafka, 0)
			if debug {
				fmt.Println(output)
			}
		}

		if tt.AuthKafka != "" {
			output := runCommand(t, ccloudTestBin, []string{}, "api-key create --resource "+tt.UseKafka, 0)
			if debug {
				fmt.Println(output)
			}
			// HACK: we don't have scriptable output yet so we parse it from the table
			key := strings.TrimSpace(strings.Split(strings.Split(output, "\n")[3], "|")[2])
			output = runCommand(t, ccloudTestBin, []string{}, fmt.Sprintf("api-key use %s --resource %s", key, tt.UseKafka), 0)
			if debug {
				fmt.Println(output)
			}
		}
		covCollectorOptions := parseCmdFuncsToCoverageCollectorOptions(tt.PreCmdFuncs, tt.PostCmdFuncs)
		output := runCommand(t, ccloudTestBin, tt.Env, tt.Args, tt.WantErrCode, covCollectorOptions...)
		if debug {
			fmt.Println(output)
		}

		if strings.HasPrefix(tt.Args, "kafka cluster create") ||
			strings.HasPrefix(tt.Args, "config context current") {
			re := regexp.MustCompile("https?://127.0.0.1:[0-9]+")
			output = re.ReplaceAllString(output, "http://127.0.0.1:12345")
		}

		s.validateTestOutput(tt, t, output)
	})
}

func (s *CLITestSuite) RunConfluentTest(tt CLITest) {
	if tt.Name == "" {
		tt.Name = tt.Args
	}
	if strings.HasPrefix(tt.Name, "error") {
		tt.WantErrCode = 1
	}
	s.T().Run(tt.Name, func(t *testing.T) {
		if !tt.Workflow {
			//t.Parallel()
			ResetConfiguration(t, "confluent")
		}

		// Executes login command if test specifies
		loginURL := s.getLoginURL("confluent", tt)
		if tt.Login == "default" {
			env := []string{"XX_CONFLUENT_USERNAME=fake@user.com", "XX_CONFLUENT_PASSWORD=pass1"}
			output := runCommand(t, confluentTestBin, env, "login --url "+loginURL, 0)
			if debug {
				fmt.Println(output)
			}
		}
		covCollectorOptions := parseCmdFuncsToCoverageCollectorOptions(tt.PreCmdFuncs, tt.PostCmdFuncs)
		output := runCommand(t, confluentTestBin, tt.Env, tt.Args, tt.WantErrCode, covCollectorOptions...)

		if strings.HasPrefix(tt.Args, "config context list") ||
			strings.HasPrefix(tt.Args, "config context current") {
			re := regexp.MustCompile("https?://127.0.0.1:[0-9]+")
			output = re.ReplaceAllString(output, "http://127.0.0.1:12345")
		}

		s.validateTestOutput(tt, t, output)
	})
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

func LoadFixture(t *testing.T, fixture string) string {
	content, err := ioutil.ReadFile(FixturePath(t, fixture))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func FixturePath(t *testing.T, fixture string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}

	return filepath.Join(filepath.Dir(filename), "fixtures", "output", fixture)
}

func GetInputFixturePath(t *testing.T, directoryName string, file string) string {
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("problems recovering caller information")
	}
	return filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", directoryName, file)
}

func ResetConfiguration(t *testing.T, cliName string) {
	// HACK: delete your current config to isolate tests cases for non-workflow tests...
	// probably don't really want to do this or devs will get mad
	cfg := v3.New(&config.Params{
		CLIName: cliName,
	})
	err := cfg.Save()
	require.NoError(t, err)
}

func binaryPath(t *testing.T, binaryName string) string {
	dir, err := os.Getwd()
	require.NoError(t, err)
	return path.Join(dir, binaryName)
}

func writeFixture(t *testing.T, fixture string, content string) {
	err := ioutil.WriteFile(FixturePath(t, fixture), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func (s *CLITestSuite) validateTestOutput(tt CLITest, t *testing.T, output string) {
	if *update && !tt.Regex && tt.Fixture != "" {
		writeFixture(t, tt.Fixture, output)
	}
	actual := utils.NormalizeNewLines(output)
	if tt.Contains != "" {
		require.Contains(t, actual, tt.Contains)
	} else if tt.NotContains != "" {
		require.NotContains(t, actual, tt.NotContains)
	} else if tt.Fixture != "" {
		expected := utils.NormalizeNewLines(LoadFixture(t, tt.Fixture))
		if tt.Regex {
			require.Regexp(t, expected, actual)
		} else if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("\n   actual:\n%s\nexpected:\n%s", actual, expected)
		}
	}
	if tt.WantFunc != nil {
		tt.WantFunc(t)
	}
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

// Parses pre and post CmdFuncs into CoverageCollectorOptions which can be unsed in covCollector.RunBinary()
func parseCmdFuncsToCoverageCollectorOptions(preCmdFuncs []bincover.PreCmdFunc, postCmdFuncs []bincover.PostCmdFunc) []bincover.CoverageCollectorOption {
	if len(preCmdFuncs) == 0 && len(postCmdFuncs) == 0 {
		return []bincover.CoverageCollectorOption{}
	}
	var options []bincover.CoverageCollectorOption
	return append(options, bincover.PreExec(preCmdFuncs...), bincover.PostExec(postCmdFuncs...))
}
