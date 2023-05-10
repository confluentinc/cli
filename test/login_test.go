package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/stretchr/testify/require"
)

var (
	urlPlaceHolder          = "<URL_PLACEHOLDER>"
	loggedInAsOutput        = fmt.Sprintf(errors.LoggedInAsMsg, "good@user.com")
	passwordPlaceholder     = "<PASSWORD_PLACEHOLDER>"
	loggedInAsWithOrgOutput = fmt.Sprintf(errors.LoggedInAsMsgWithOrg, "good@user.com", "abc-123", "Confluent")
	loggedInEnvOutput       = fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default")
)

func (s *CLITestSuite) TestCcloudLoginUseKafkaAuthKafkaErrors() {
	tests := []CLITest{
		{
			name:        "error if no active kafka",
			args:        "kafka topic create integ",
			fixture:     "login/err-no-kafka.golden",
			wantErrCode: 1,
			login:       "default",
		},
		{
			name:        "error if topic already exists",
			args:        "kafka topic create topic-exist",
			fixture:     "login/topic-exists.golden",
			wantErrCode: 1,
			login:       "default",
			useKafka:    "lkc-create-topic",
			authKafka:   "true",
			env:         []string{"XX_CCLOUD_USE_KAFKA_REST=true"},
		},
		{
			name:        "error if no api key used",
			args:        "kafka topic produce integ",
			fixture:     "login/err-no-api-key.golden",
			wantErrCode: 1,
			login:       "default",
			useKafka:    "lkc-abc123",
		},
		{
			name:        "error if deleting non-existent api-key",
			args:        "api-key delete UNKNOWN",
			fixture:     "login/delete-unknown-key.golden",
			wantErrCode: 1,
			login:       "default",
			useKafka:    "lkc-abc123",
			authKafka:   "true",
		},
		{
			name:        "error if using unknown kafka",
			args:        "kafka cluster use lkc-unknown",
			fixture:     "login/err-use-unknown-kafka.golden",
			wantErrCode: 1,
			login:       "default",
		},
	}

	for _, tt := range tests {
		s.runCcloudTest(tt)
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
		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0)
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
		wantBytes, err := ioutil.ReadFile(wantFile)
		s.NoError(err)
		want := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, -1)
		data := v1.Config{}
		err = json.Unmarshal(got, &data)
		s.NoError(err)
		want = strings.Replace(want, passwordPlaceholder, data.SavedCredentials["login-good@user.com-"+tt.loginURL].EncryptedPassword, -1)
		require.Contains(s.T(), utils.NormalizeNewLines(string(got)), utils.NormalizeNewLines(string(want)))
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
			env = []string{fmt.Sprintf("%s=good@user.com", pauth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", pauth.ConfluentCloudPassword)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", pauth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", pauth.ConfluentPlatformPassword)}
		}

		configFile := filepath.Join(os.Getenv("HOME"), ".confluent", "config.json")
		old, err := os.ReadFile(configFile)
		s.NoError(err)
		oldData := v1.Config{}
		err = json.Unmarshal(old, &oldData)
		s.NoError(err)

		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0)
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
			name:        "invalid URL provided",
			args:        "login --url http:///test",
			fixture:     "login/invalid-login-url.golden",
			wantErrCode: 1,
		},
	}

	for _, tt := range tests {
		tt.loginURL = s.TestBackend.GetMdsUrl()
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestLogin_CaCertPath() {
	resetConfiguration(s.T())

	tests := []CLITest{
		{args: fmt.Sprintf("login --url %s --ca-cert-path test/fixtures/input/login/test.crt", s.TestBackend.GetMdsUrl())},
		{args: "context list -o yaml", fixture: "login/1.golden", regex: true},
	}

	env := []string{
		fmt.Sprintf("%s=%s", pauth.ConfluentPlatformUsername, "on-prem@example.com"),
		fmt.Sprintf("%s=%s", pauth.ConfluentPlatformPassword, "password"),
	}

	for _, tt := range tests {
		out := runCommand(s.T(), testBin, env, tt.args, 0)
		s.validateTestOutput(tt, s.T(), out)
	}
}
