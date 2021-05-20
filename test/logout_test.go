package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	RemovedFromNetrcOutput = "Removed credentials for user \"good@user.com\" from netrc file \"/tmp/netrc_test\""
	loggedOutOutput        = fmt.Sprintf(errors.LoggedOutMsg)
)

func (s *CLITestSuite) TestRemoveUsernamePassword() {
	type saveTest struct {
		input    string
		cliName  string
		want     string
		loginURL string
		bin      string
	}
	cloudUrl := s.TestBackend.GetCloudUrl()
	mdsUrl := s.TestBackend.GetMdsUrl()
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	tests := []saveTest{
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove-ccloud"),
			"ccloud",
			"netrc-remove-username-password.golden",
			cloudUrl,
			ccloudTestBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove-mds"),
			"confluent",
			"netrc-remove-username-password.golden",
			mdsUrl,
			confluentTestBin,
		},
	}
	for _, tt := range tests {
		// store existing credentials in a temp netrc to check that they are not corrupted
		var env []string
		if tt.cliName == "ccloud" {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", auth.CCloudPasswordEnvVar)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentUsernameEnvVar), fmt.Sprintf("%s=pass1", auth.ConfluentPasswordEnvVar)}
		}
		originalNetrc, err := ioutil.ReadFile(tt.input)
		s.NoError(err)
		original := strings.Replace(string(originalNetrc), urlPlaceHolder, tt.loginURL, 1)
		err = ioutil.WriteFile(netrc.NetrcIntegrationTestFile, []byte(original), 0600)
		s.NoError(err)

		// run login to provide context, then logout command and check output
		runCommand(s.T(), tt.bin, env, "login --url "+tt.loginURL, 0)
		output := runCommand(s.T(), tt.bin, env, "logout -vvvv", 0)
		s.Contains(output, loggedOutOutput)
		s.Contains(output, RemovedFromNetrcOutput)

		// check netrc file doesn't contain credentials
		got, err := ioutil.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantFile := filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", tt.want)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(wantFile)
		s.NoError(err)
		fmt.Println("Here ------------------------------")
		fmt.Println(string(wantBytes))
		fmt.Println(string(got))
		s.Equal(utils.NormalizeNewLines(string(wantBytes)), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}
