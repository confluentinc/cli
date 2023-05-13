package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	test_server "github.com/confluentinc/cli/test/test-server"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/auth"
	pauth "github.com/confluentinc/cli/internal/pkg/auth"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	urlPlaceHolder      = "<URL_PLACEHOLDER>"
	passwordPlaceholder = "<PASSWORD_PLACEHOLDER>"
	loggedInAsOutput    = fmt.Sprintf(errors.LoggedInAsMsg, "good@user.com")
	loggedInEnvOutput   = fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default")
)

func (s *CLITestSuite) TestCcloudLoginUseKafkaAuthKafkaErrors() {
	tests := []CLITest{
		{
			name:        "error if not authenticated",
			args:        "kafka topic create integ",
			fixture:     "err-not-authenticated.golden",
			wantErrCode: 1,
		},
		{
			name:        "error if no active kafka",
			args:        "kafka topic create integ",
			fixture:     "err-no-kafka.golden",
			wantErrCode: 1,
			login:       "default",
		},
		{
			name:        "error if topic already exists",
			args:        "kafka topic create topic-exist",
			fixture:     "topic-exists.golden",
			wantErrCode: 1,
			login:       "default",
			useKafka:    "lkc-create-topic",
			authKafka:   "true",
			env:         []string{"XX_CCLOUD_USE_KAFKA_REST=true"},
		},
		{
			name:        "error if no api key used",
			args:        "kafka topic produce integ",
			fixture:     "err-no-api-key.golden",
			wantErrCode: 1,
			login:       "default",
			useKafka:    "lkc-abc123",
		},
		{
			name:        "error if deleting non-existent api-key",
			args:        "api-key delete UNKNOWN",
			fixture:     "delete-unknown-key.golden",
			wantErrCode: 1,
			login:       "default",
			useKafka:    "lkc-abc123",
			authKafka:   "true",
		},
		{
			name:        "error if using unknown kafka",
			args:        "kafka cluster use lkc-unknown",
			fixture:     "err-use-unknown-kafka.golden",
			wantErrCode: 1,
			login:       "default",
		},
	}

	for _, tt := range tests {
		s.runCcloudTest(tt)
	}
}

func serveCloudBackend(t *testing.T) *test_server.TestBackend {
	router := test_server.NewCloudRouter(t)
	return test_server.NewCloudTestBackendFromRouters(router, test_server.NewEmptyKafkaRouter())
}

func serveMDSBackend(t *testing.T) *test_server.TestBackend {
	router := test_server.NewMdsRouter(t)
	return test_server.NewConfluentTestBackendFromRouter(router)
}

func (s *CLITestSuite) TestSaveUsernamePassword() {
	type saveTest struct {
		cliName  string
		want     string
		loginURL string
		bin      string
	}
	cloudBackend := serveCloudBackend(s.T())
	defer cloudBackend.Close()
	mdsServer := serveMDSBackend(s.T())
	defer mdsServer.Close()
	tests := []saveTest{
		{
			"ccloud",
			"config-save-ccloud-username-password.golden",
			cloudBackend.GetCloudUrl(),
			ccloudTestBin,
		},
		{
			"confluent",
			"config-save-mds-username-password.golden",
			mdsServer.GetMdsUrl(),
			confluentTestBin,
		},
	}
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	for _, tt := range tests {
		configFile := filepath.Join(os.Getenv("HOME"), "."+tt.cliName, "config.json")
		// run the login command with --save flag and check output
		var env []string
		if tt.cliName == "ccloud" {
			env = []string{fmt.Sprintf("%s=good@user.com", pauth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", auth.CCloudPasswordEnvVar)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", pauth.ConfluentUsernameEnvVar), fmt.Sprintf("%s=pass1", auth.ConfluentPasswordEnvVar)}
		}
		_ = runCommand(s.T(), tt.bin, env, "logout -vvvv", 0)
		//TODO add save test using stdin input
		output := runCommand(s.T(), tt.bin, env, "login --save -vvv --url "+tt.loginURL, 0)
		if tt.cliName == "ccloud" {
			s.Contains(output, loggedInEnvOutput)
		}

		got, err := os.ReadFile(configFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(wantFile)
		s.NoError(err)
		want := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, -1)
		data := v3.Config{}
		err = json.Unmarshal(got, &data)
		s.NoError(err)
		fmt.Println("login-good@user.com-" + tt.loginURL)
		fmt.Printf("%+v\n", data)
		want = strings.Replace(want, passwordPlaceholder, data.SavedCredentials["login-good@user.com-"+tt.loginURL].EncryptedPassword, -1)
		require.Contains(s.T(), utils.NormalizeNewLines(string(got)), utils.NormalizeNewLines(string(want)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestMDSLoginURL() {
	tests := []CLITest{
		{
			name:        "invalid URL provided",
			args:        "login --url http:///test",
			fixture:     "invalid-login-url.golden",
			wantErrCode: 1,
		},
	}
	mdsServer := serveMDSBackend(s.T())
	defer mdsServer.Close()

	for _, tt := range tests {
		tt.loginURL = mdsServer.GetMdsUrl()
		s.runConfluentTest(tt)
	}
}
