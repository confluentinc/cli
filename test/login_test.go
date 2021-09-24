package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	test_server "github.com/confluentinc/cli/test/test-server"

	"github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	urlPlaceHolder     = "<URL_PLACEHOLDER>"
	savedToNetrcOutput = fmt.Sprintf(errors.WroteCredentialsToNetrcMsg, "/tmp/netrc_test")
	loggedInAsOutput   = fmt.Sprintf(errors.LoggedInAsMsg, "good@user.com")
	loggedInEnvOutput  = fmt.Sprintf(errors.LoggedInUsingEnvMsg, "a-595", "default")
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
			"netrc-save-ccloud-username-password.golden",
			cloudBackend.GetCloudUrl(),
			ccloudTestBin,
		},
		{
			"confluent",
			"netrc-save-mds-username-password.golden",
			mdsServer.GetMdsUrl(),
			confluentTestBin,
		},
	}
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	netrcInput := filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc")
	for _, tt := range tests {
		// store existing credentials in netrc to check that they are not corrupted
		originalNetrc, err := ioutil.ReadFile(netrcInput)
		s.NoError(err)
		err = ioutil.WriteFile(netrc.NetrcIntegrationTestFile, originalNetrc, 0600)
		s.NoError(err)

		// run the login command with --save flag and check output
		var env []string
		if tt.cliName == "ccloud" {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", auth.CCloudPasswordEnvVar)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentUsernameEnvVar), fmt.Sprintf("%s=pass1", auth.ConfluentPasswordEnvVar)}
		}
		//TODO add save test using stdin input
		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0)
		s.Contains(output, savedToNetrcOutput)
		s.Contains(output, loggedInAsOutput)
		if tt.cliName == "ccloud" {
			s.Contains(output, loggedInEnvOutput)
		}

		// check netrc file result
		got, err := ioutil.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(wantFile)
		s.NoError(err)
		want := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, 1)
		s.Equal(utils.NormalizeNewLines(want), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestUpdateNetrcPassword() {
	type updateTest struct {
		input    string
		cliName  string
		want     string
		loginURL string
		bin      string
	}
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	cloudServer := serveCloudBackend(s.T())
	defer cloudServer.Close()
	mdsServer := serveMDSBackend(s.T())
	defer mdsServer.Close()
	tests := []updateTest{
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-old-password-ccloud"),
			"ccloud",
			"netrc-save-ccloud-username-password.golden",
			cloudServer.GetCloudUrl(),
			ccloudTestBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-old-password-mds"),
			"confluent",
			"netrc-save-mds-username-password.golden",
			mdsServer.GetMdsUrl(),
			confluentTestBin,
		},
	}
	for _, tt := range tests {
		// store existing credential + the user credential to be updated
		originalNetrc, err := ioutil.ReadFile(tt.input)
		s.NoError(err)
		originalNetrcString := strings.Replace(string(originalNetrc), urlPlaceHolder, tt.loginURL, 1)
		err = ioutil.WriteFile(netrc.NetrcIntegrationTestFile, []byte(originalNetrcString), 0600)
		s.NoError(err)

		// run the login command with --save flag and check output
		var env []string
		if tt.cliName == "ccloud" {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", auth.CCloudPasswordEnvVar)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentUsernameEnvVar), fmt.Sprintf("%s=pass1", auth.ConfluentPasswordEnvVar)}
		}
		output := runCommand(s.T(), tt.bin, env, "login -vvv --save --url "+tt.loginURL, 0)
		s.Contains(output, savedToNetrcOutput)
		s.Contains(output, loggedInAsOutput)
		if tt.cliName == "ccloud" {
			s.Contains(output, loggedInEnvOutput)
		}

		// check netrc file result
		got, err := ioutil.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(wantFile)
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
