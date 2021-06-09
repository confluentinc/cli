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
	removedFromNetrcOutput = "Removed credentials for user \"good@user.com\" from netrc file \"/tmp/netrc_test\""
	loggedOutOutput        = fmt.Sprintf(errors.LoggedOutMsg)
)

func (s *CLITestSuite) TestRemoveUsernamePassword() {
	type saveTest struct {
		input    string
		cliName  string
		wantFile string
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
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove"),
			"ccloud",
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", "netrc-remove-username-password.golden"),
			cloudUrl,
			ccloudTestBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove"),
			"confluent",
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", "netrc-remove-username-password.golden"),
			mdsUrl,
			confluentTestBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-empty"),
			"ccloud",
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", "empty.golden"),
			cloudUrl,
			ccloudTestBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-empty"),
			"confluent",
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", "empty.golden"),
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
		err = ioutil.WriteFile(netrc.NetrcIntegrationTestFile, originalNetrc, 0600)
		s.NoError(err)

		// run login to provide context, then logout command and check output
		output := runCommand(s.T(), tt.bin, env, "login -vvvv --save --url "+tt.loginURL, 0)
		s.Contains(output, savedToNetrcOutput)
		s.Contains(output, loggedInAsOutput)

		output = runCommand(s.T(), tt.bin, env, "logout -vvvv", 0)
		s.Contains(output, loggedOutOutput)
		s.Contains(output, removedFromNetrcOutput)

		// check netrc file matches wanted file
		got, err := ioutil.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(tt.wantFile)
		s.NoError(err)
		s.Equal(utils.NormalizeNewLines(string(wantBytes)), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestRemoveUsernamePasswordFail() {
	// fail to parse the netrc file should leave it unchanged
	type saveTest struct {
		input    string
		cliName  string
		wantFile string
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
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove-ccloud-fail"),
			"ccloud",
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove-ccloud-fail"),
			cloudUrl,
			ccloudTestBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove-mds-fail"),
			"confluent",
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove-mds-fail"),
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
		runCommand(s.T(), tt.bin, env, "login --url "+tt.loginURL, 0) // without save flag so the netrc file won't be modified
		output := runCommand(s.T(), tt.bin, env, "logout", 0)
		s.Contains(output, loggedOutOutput)

		// check netrc file matches wanted file
		got, err := ioutil.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantBytes, err := ioutil.ReadFile(tt.wantFile)
		s.NoError(err)
		wantString := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, 1)
		s.Equal(utils.NormalizeNewLines(wantString), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}
