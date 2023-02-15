package test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/confluentinc/bincover"
	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/auth"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	urlPlaceHolder          = "<URL_PLACEHOLDER>"
	savedToNetrcOutput      = fmt.Sprintf(errors.WroteCredentialsToNetrcMsg, "netrc_test")
	loggedInAsOutput        = fmt.Sprintf(errors.LoggedInAsMsg, "good@user.com")
	loggedInAsWithOrgOutput = fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "good@user.com", "abc-123", "Confluent")
	loggedInEnvOutput       = fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default")
)

func (s *CLITestSuite) TestLogin_Help() {
	s.runIntegrationTest(CLITest{args: "login -h", fixture: "login/help.golden"})
}

func (s *CLITestSuite) TestLogin_VariousOrgSuspensionStatus() {
	args := fmt.Sprintf("login --url %s -vvv", s.TestBackend.GetCloudUrl())

	s.T().Run("free trial organization login", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}
		os.Setenv("IS_ON_FREE_TRIAL", "true")
		defer unsetFreeTrialEnv()

		output := runCommand(tt, testBin, env, args, 0)
		require.Contains(tt, output, loggedInAsWithOrgOutput)
		require.Contains(tt, output, loggedInEnvOutput)
		require.Contains(tt, output, fmt.Sprintf(errors.RemainingFreeCreditMsg, 40.00))
	})

	s.T().Run("non-free-trial organization login", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}

		output := runCommand(tt, testBin, env, args, 0)
		require.Contains(tt, output, loggedInAsWithOrgOutput)
		require.Contains(tt, output, loggedInEnvOutput)
		require.NotContains(tt, output, "Free credits")
	})

	s.T().Run("suspended organization login", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=suspended@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(tt, testBin, env, args, 1)
		require.Contains(tt, output, new(ccloudv1.SuspendedOrganizationError).Error())
		require.Contains(tt, output, errors.SuspendedOrganizationSuggestions)
	})

	s.T().Run("end of free trial suspended organization", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=end-of-free-trial-suspended@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(tt, testBin, env, args, 0)
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "end-of-free-trial-suspended@user.com", "abc-123", "Confluent"))
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default"))
		require.Contains(tt, output, fmt.Sprintf(errors.EndOfFreeTrialErrorMsg, "test-org"))
	})
}

func (s *CLITestSuite) TestCcloudErrors() {
	args := fmt.Sprintf("login --url %s -vvv", s.TestBackend.GetCloudUrl())

	s.T().Run("invalid user or pass", func(tt *testing.T) {
		env := []string{fmt.Sprintf("%s=incorrect@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(tt, testBin, env, args, 1)
		require.Contains(tt, output, errors.InvalidLoginErrorMsg)
		require.Contains(tt, output, errors.ComposeSuggestionsMessage(errors.InvalidLoginErrorSuggestions))
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
		output := runCommand(tt, testBin, env, args, 0)
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "invalid@user.com", "abc-123", "Confluent"))
		require.Contains(tt, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default"))

		output = runCommand(s.T(), testBin, []string{}, "kafka cluster list", 1)
		require.Contains(tt, output, errors.CorruptedTokenErrorMsg)
		require.Contains(tt, output, errors.ComposeSuggestionsMessage(errors.CorruptedTokenSuggestions))
	})
}

func (s *CLITestSuite) TestCcloudLoginUseKafkaAuthKafkaErrors() {
	tests := []CLITest{
		{
			name:        "error if no active kafka",
			args:        "kafka topic create integ",
			fixture:     "login/err-no-kafka.golden",
			wantErrCode: 1,
		},
		{
			name:        "error if topic already exists",
			args:        "kafka topic create topic-exist",
			fixture:     "login/topic-exists.golden",
			wantErrCode: 1,
			useKafka:    "lkc-create-topic",
			authKafka:   "true",
		},
		{
			name:        "error if no API key used",
			args:        "kafka topic produce integ",
			fixture:     "login/err-no-api-key.golden",
			wantErrCode: 1,
			useKafka:    "lkc-abc123",
		},
		{
			name:        "error if deleting non-existent api-key",
			args:        "api-key delete UNKNOWN",
			fixture:     "login/delete-unknown-key.golden",
			wantErrCode: 1,
			useKafka:    "lkc-abc123",
			authKafka:   "true",
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("y\n"))},
		},
		{
			name:        "error if using unknown kafka",
			args:        "kafka cluster use lkc-unknown",
			fixture:     "login/err-use-unknown-kafka.golden",
			wantErrCode: 1,
		},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestSaveUsernamePassword() {
	tests := []struct {
		isCloud  bool
		want     string
		loginURL string
		bin      string
	}{
		{
			true,
			"login/netrc-save-ccloud-username-password.golden",
			s.TestBackend.GetCloudUrl(),
			testBin,
		},
		{
			false,
			"login/netrc-save-mds-username-password.golden",
			s.TestBackend.GetMdsUrl(),
			testBin,
		},
	}

	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}

	netrcInput := filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "login", "netrc")
	for _, tt := range tests {
		// store existing credentials in netrc to check that they are not corrupted
		originalNetrc, err := os.ReadFile(netrcInput)
		s.NoError(err)
		err = os.WriteFile(netrc.NetrcIntegrationTestFile, originalNetrc, 0600)
		s.NoError(err)

		// run the login command with --save flag and check output
		var env []string
		if tt.isCloud {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", auth.ConfluentPlatformPassword)}
		}

		// TODO: add save test using stdin input
		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0)
		s.Contains(output, savedToNetrcOutput)
		if tt.isCloud {
			s.Contains(output, loggedInAsWithOrgOutput)
			s.Contains(output, loggedInEnvOutput)
		} else {
			s.Contains(output, loggedInAsOutput)
		}

		// check netrc file result
		got, err := os.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := os.ReadFile(wantFile)
		s.NoError(err)
		want := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, 1)
		s.Equal(utils.NormalizeNewLines(want), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestUpdateNetrcPassword() {
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}

	tests := []struct {
		input    string
		isCloud  bool
		want     string
		loginURL string
		bin      string
	}{
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "login", "netrc-old-password-ccloud"),
			true,
			"login/netrc-save-ccloud-username-password.golden",
			s.TestBackend.GetCloudUrl(),
			testBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "login", "netrc-old-password-mds"),
			false,
			"login/netrc-save-mds-username-password.golden",
			s.TestBackend.GetMdsUrl(),
			testBin,
		},
	}

	for _, tt := range tests {
		// store existing credential + the user credential to be updated
		originalNetrc, err := os.ReadFile(tt.input)
		s.NoError(err)
		originalNetrcString := strings.Replace(string(originalNetrc), urlPlaceHolder, tt.loginURL, 1)
		err = os.WriteFile(netrc.NetrcIntegrationTestFile, []byte(originalNetrcString), 0600)
		s.NoError(err)

		// run the login command with --save flag and check output
		var env []string
		if tt.isCloud {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", auth.ConfluentPlatformPassword)}
		}
		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0)
		s.Contains(output, savedToNetrcOutput)
		if tt.isCloud {
			s.Contains(output, loggedInAsWithOrgOutput)
			s.Contains(output, loggedInEnvOutput)
		} else {
			s.Contains(output, loggedInAsOutput)
		}

		// check netrc file result
		got, err := os.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := os.ReadFile(wantFile)
		s.NoError(err)
		want := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, 1)
		s.Equal(utils.NormalizeNewLines(want), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestMDSLoginURL() {
	tests := []CLITest{
		{
			name:        "invalid URL provided",
			args:        "login --url http:///test",
			fixture:     "login/invalid-login-url.golden",
			wantErrCode: 1,
		},
	}

	for _, tt := range tests {
		tt.loginURL = s.TestBackend.GetMdsUrl()
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestLogin_CaCertPath() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{
			env:  []string{"CONFLUENT_PLATFORM_USERNAME=on-prem@example.com", "CONFLUENT_PLATFORM_PASSWORD=password"},
			args: fmt.Sprintf("login --url %s --ca-cert-path test/fixtures/input/login/test.crt", s.TestBackend.GetMdsUrl()),
		},
		{
			args:    "context list -o yaml",
			fixture: "login/1.golden",
			regex:   true,
		},
	}

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestLogin_SsoCodeInvalidFormat() {
	resetConfiguration(s.T(), false)

	tt := CLITest{
		env:         []string{"CONFLUENT_CLOUD_EMAIL=sso@test.com"},
		args:        fmt.Sprintf("login --url %s --no-browser", s.TestBackend.GetCloudUrl()),
		fixture:     "login/sso.golden",
		regex:       true,
		wantErrCode: 1,
	}

	// TODO: Accept text input in integration tests

	s.runIntegrationTest(tt)
}
