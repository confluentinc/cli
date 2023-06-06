package test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/auth"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	urlPlaceHolder          = "<URL_PLACEHOLDER>"
	passwordPlaceholder     = "<PASSWORD_PLACEHOLDER>"
	loggedInAsOutput        = fmt.Sprintf(errors.LoggedInAsMsg, "good@user.com")
	loggedInAsWithOrgOutput = fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "good@user.com", "abc-123", "Confluent")
	loggedInEnvOutput       = fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595")
)

func (s *CLITestSuite) TestLogin_VariousOrgSuspensionStatus() {
	args := fmt.Sprintf("login --url %s -vvv", s.TestBackend.GetCloudUrl())

	s.T().Run("free trial organization login", func(t *testing.T) {
		env := []string{fmt.Sprintf("%s=good@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		t.Setenv("IS_ON_FREE_TRIAL", "true")

		output := runCommand(t, testBin, env, args, 0, "")
		require.Contains(t, output, loggedInAsWithOrgOutput)
		require.Contains(t, output, loggedInEnvOutput)
		require.Contains(t, output, fmt.Sprintf(errors.RemainingFreeCreditMsg, 40.00))
	})

	s.T().Run("non-free-trial organization login", func(t *testing.T) {
		env := []string{fmt.Sprintf("%s=good@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}

		output := runCommand(t, testBin, env, args, 0, "")
		require.Contains(t, output, loggedInAsWithOrgOutput)
		require.Contains(t, output, loggedInEnvOutput)
		require.NotContains(t, output, "Free credits")
	})

	s.T().Run("suspended organization login", func(t *testing.T) {
		env := []string{fmt.Sprintf("%s=suspended@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(t, testBin, env, args, 1, "")
		require.Contains(t, output, new(ccloudv1.SuspendedOrganizationError).Error())
		require.Contains(t, output, errors.SuspendedOrganizationSuggestions)
	})

	s.T().Run("end of free trial suspended organization", func(t *testing.T) {
		env := []string{fmt.Sprintf("%s=end-of-free-trial-suspended@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(t, testBin, env, args, 0, "")
		require.Contains(t, output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "end-of-free-trial-suspended@user.com", "abc-123", "Confluent"))
		require.Contains(t, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595"))
		require.Contains(t, output, fmt.Sprintf(errors.EndOfFreeTrialErrorMsg, "test-org"))
	})
}

func (s *CLITestSuite) TestCcloudErrors() {
	args := fmt.Sprintf("login --url %s -vvv", s.TestBackend.GetCloudUrl())

	s.T().Run("invalid user or pass", func(t *testing.T) {
		env := []string{fmt.Sprintf("%s=incorrect@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(t, testBin, env, args, 1, "")
		require.Contains(t, output, errors.InvalidLoginErrorMsg)
		require.Contains(t, output, errors.ComposeSuggestionsMessage(errors.InvalidLoginErrorSuggestions))
	})

	s.T().Run("expired token", func(t *testing.T) {
		env := []string{fmt.Sprintf("%s=expired@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(t, testBin, env, args, 0, "")
		require.Contains(t, output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "expired@user.com", "abc-123", "Confluent"))
		require.Contains(t, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595"))
		output = runCommand(t, testBin, []string{}, "kafka cluster list", 1, "")
		require.Contains(t, output, errors.ExpiredTokenErrorMsg)
		require.Contains(t, output, errors.ComposeSuggestionsMessage(errors.ExpiredTokenSuggestions))
	})

	s.T().Run("malformed token", func(t *testing.T) {
		env := []string{fmt.Sprintf("%s=malformed@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(t, testBin, env, "logout", 0, "")
		require.Contains(t, output, errors.LoggedOutMsg)
		output = runCommand(t, testBin, env, args, 0, "")
		require.Contains(t, output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "malformed@user.com", "abc-123", "Confluent"))
		require.Contains(t, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595"))

		output = runCommand(t, testBin, []string{}, "kafka cluster list", 1, "")
		require.Contains(t, output, errors.CorruptedTokenErrorMsg)
		require.Contains(t, output, errors.ComposeSuggestionsMessage(errors.CorruptedTokenSuggestions))
	})

	s.T().Run("invalid jwt", func(t *testing.T) {
		env := []string{fmt.Sprintf("%s=invalid@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		output := runCommand(t, testBin, env, "logout", 0, "")
		require.Contains(t, output, errors.LoggedOutMsg)
		output = runCommand(t, testBin, env, args, 0, "")
		require.Contains(t, output, fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "invalid@user.com", "abc-123", "Confluent"))
		require.Contains(t, output, fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595"))

		output = runCommand(t, testBin, []string{}, "kafka cluster list", 1, "")
		require.Contains(t, output, errors.CorruptedTokenErrorMsg)
		require.Contains(t, output, errors.ComposeSuggestionsMessage(errors.CorruptedTokenSuggestions))
	})
}

func (s *CLITestSuite) TestCcloudLoginUseKafkaAuthKafkaErrors() {
	tests := []CLITest{
		{
			name:     "error if no active kafka",
			args:     "kafka topic create integ",
			fixture:  "login/err-no-kafka.golden",
			exitCode: 1,
		},
		{
			name:      "error if topic already exists",
			args:      "kafka topic create topic-exist",
			fixture:   "login/topic-exists.golden",
			exitCode:  1,
			useKafka:  "lkc-create-topic",
			authKafka: "true",
		},
		{
			name:     "error if no API key used",
			args:     "kafka topic produce integ",
			fixture:  "login/err-no-api-key.golden",
			exitCode: 1,
			useKafka: "lkc-abc123",
		},
		{
			name:      "error if deleting non-existent api-key",
			args:      "api-key delete UNKNOWN",
			input:     "y\n",
			fixture:   "login/delete-unknown-key.golden",
			exitCode:  1,
			useKafka:  "lkc-abc123",
			authKafka: "true",
		},
		{
			name:     "error if using unknown kafka",
			args:     "kafka cluster use lkc-unknown",
			fixture:  "login/err-use-unknown-kafka.golden",
			exitCode: 1,
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
			"login/config-save-ccloud-username-password.golden",
			s.TestBackend.GetCloudUrl(),
			testBin,
		},
		{
			false,
			"login/config-save-mds-username-password.golden",
			s.TestBackend.GetMdsUrl(),
			testBin,
		},
	}

	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}

	for _, tt := range tests {
		configFile := filepath.Join(os.Getenv("HOME"), ".confluent", "config.json")
		// run the login command with --save flag and check output
		var env []string
		if tt.isCloud {
			env = []string{fmt.Sprintf("%s=good@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", pauth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", pauth.ConfluentPlatformPassword)}
		}

		// TODO: add save test using stdin input
		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0, "")
		if tt.isCloud {
			s.Contains(output, loggedInAsWithOrgOutput)
			s.Contains(output, loggedInEnvOutput)
		} else {
			s.Contains(output, loggedInAsOutput)
		}

		// check netrc file result
		got, err := os.ReadFile(configFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := os.ReadFile(wantFile)
		s.NoError(err)
		want := strings.ReplaceAll(string(wantBytes), urlPlaceHolder, tt.loginURL)
		data := v1.Config{}
		err = json.Unmarshal(got, &data)
		s.NoError(err)
		want = strings.ReplaceAll(want, passwordPlaceholder, data.SavedCredentials["login-good@user.com-"+tt.loginURL].EncryptedPassword)
		require.Contains(s.T(), utils.NormalizeNewLines(string(got)), utils.NormalizeNewLines(want))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestUpdateNetrcPassword() {
	tests := []struct {
		isCloud  bool
		loginURL string
		bin      string
	}{
		{
			true,
			s.TestBackend.GetCloudUrl(),
			testBin,
		},
		{
			false,
			s.TestBackend.GetMdsUrl(),
			testBin,
		},
	}

	for _, tt := range tests {
		// run the login command with --save flag and check output
		var env []string
		if tt.isCloud {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", auth.ConfluentPlatformPassword)}
		}

		configFile := filepath.Join(os.Getenv("HOME"), ".confluent", "config.json")
		old, err := os.ReadFile(configFile)
		s.NoError(err)
		oldData := v1.Config{}
		err = json.Unmarshal(old, &oldData)
		s.NoError(err)

		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0, "")
		if tt.isCloud {
			s.Contains(output, loggedInAsWithOrgOutput)
			s.Contains(output, loggedInEnvOutput)
		} else {
			s.Contains(output, loggedInAsOutput)
		}

		got, err := os.ReadFile(configFile)
		s.NoError(err)
		data := v1.Config{}
		err = json.Unmarshal(got, &data)
		s.NoError(err)

		s.NotEqual(oldData.SavedCredentials, data.SavedCredentials)
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestMDSLoginURL() {
	tests := []CLITest{
		{
			name:     "invalid URL provided",
			args:     "login --url http:///test",
			fixture:  "login/invalid-login-url.golden",
			exitCode: 1,
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
		env:      []string{"CONFLUENT_CLOUD_EMAIL=sso@test.com"},
		args:     fmt.Sprintf("login --url %s --no-browser", s.TestBackend.GetCloudUrl()),
		fixture:  "login/sso.golden",
		regex:    true,
		exitCode: 1,
	}

	// TODO: Accept text input in integration tests

	s.runIntegrationTest(tt)
}
