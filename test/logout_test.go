package test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/netrc"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	loggedOutOutput = fmt.Sprintf(errors.LoggedOutMsg)
)

func (s *CLITestSuite) TestRemoveUsernamePassword() {
	type saveTest struct {
		cliName  string
		loginURL string
		bin      string
	}
	cloudUrl := s.TestBackend.GetCloudUrl()
	mdsUrl := s.TestBackend.GetMdsUrl()
	tests := []saveTest{
		{
			"ccloud",
			cloudUrl,
			ccloudTestBin,
		},
		{
			"confluent",
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

		configFile := filepath.Join(os.Getenv("HOME"), "."+tt.cliName, "config.json")
		// run login to provide context, then logout command and check output
		output := runCommand(s.T(), tt.bin, env, "login -vvvv --save --url "+tt.loginURL, 0)
		s.Contains(output, loggedInAsOutput)

		got, err := os.ReadFile(configFile)
		s.NoError(err)
		s.Require().Contains(utils.NormalizeNewLines(string(got)), "saved_credentials")

		output = runCommand(s.T(), tt.bin, env, "logout -vvvv", 0)
		s.Contains(output, loggedOutOutput)
		got, err = os.ReadFile(configFile)
		s.NoError(err)
		s.Require().NotContains(utils.NormalizeNewLines(string(got)), `"url": `+tt.loginURL)
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestUsernamePasswordFail() {
	// fail to parse the netrc file should leave it unchanged
	type saveTest struct {
		cliName  string
		loginURL string
		bin      string
	}
	cloudUrl := s.TestBackend.GetCloudUrl()
	mdsUrl := s.TestBackend.GetMdsUrl()
	tests := []saveTest{
		{
			"ccloud",
			cloudUrl,
			ccloudTestBin,
		},
		{
			"confluent",
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

		configFile := filepath.Join(os.Getenv("HOME"), ".confluent", "config.json")
		got, err := os.ReadFile(configFile)
		s.NoError(err)
		s.Require().Contains(utils.NormalizeNewLines(string(got)), "saved_credentials")

		// run login to provide context, then logout command and check output
		runCommand(s.T(), tt.bin, env, "login --url "+tt.loginURL, 0) // without save flag so the netrc file won't be modified
		output := runCommand(s.T(), tt.bin, env, "logout", 0)
		s.Contains(output, loggedOutOutput)

		got, err = os.ReadFile(configFile)
		s.NoError(err)
		s.Require().NotContains(utils.NormalizeNewLines(string(got)), `"url": `+tt.loginURL)
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}
