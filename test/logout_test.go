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
	RemovedFromNetrcOutput = "Removed credentials for user \"login-cli-mock-email@confluent.io\" from netrc file \"/tmp/netrc_test\""
	loggedOutOutput        = fmt.Sprintf(errors.LoggedOutMsg)
)

func (s *CLITestSuite) TestRemoveUsernamePassword() {
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
			"netrc-remove-ccloud-username-password.golden",
			cloudBackend.GetCloudUrl(),
			ccloudTestBin,
		},
		{
			"confluent",
			"netrc-remove-username-password.golden",
			mdsServer.GetMdsUrl(),
			confluentTestBin,
		},
	}
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		s.T().Fatalf("problems recovering caller information")
	}
	var netrcInput string
	for _, tt := range tests {
		// store existing credentials in a temp netrc to check that they are not corrupted
		var env []string
		if tt.cliName == "ccloud" {
			env = []string{fmt.Sprintf("%s=cli-mock-email@confluent.io", auth.CCloudEmailEnvVar), fmt.Sprintf("%s=pass1", auth.CCloudPasswordEnvVar)}
			netrcInput = filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove-ccloud")
		} else {
			env = []string{fmt.Sprintf("%s=cli-mock-email@confluent.io", auth.ConfluentUsernameEnvVar), fmt.Sprintf("%s=pass1", auth.ConfluentPasswordEnvVar)}
			netrcInput = filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "netrc-remove-mds")
		}
		originalNetrc, err := ioutil.ReadFile(netrcInput)
		s.NoError(err)
		err = ioutil.WriteFile(netrc.NetrcIntegrationTestFile, originalNetrc, 0600)
		s.NoError(err)

		// run logout command and check output
		output := runCommand(s.T(), tt.bin, env, "logout -vvvv", 0)
		fmt.Println("output of logout is: ", output)
		s.Contains(output, loggedOutOutput)
		s.Contains(output, RemovedFromNetrcOutput)

		// check netrc file doesn't contain credentials
		got, err := ioutil.ReadFile(netrc.NetrcIntegrationTestFile)
		fmt.Println("got is: ", string(got))
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
