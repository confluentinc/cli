package test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/confluentinc/cli/internal/pkg/auth"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (s *CLITestSuite) TestRemoveUsernamePassword() {
	type saveTest struct {
		isCloud  bool
		loginURL string
		bin      string
	}
	cloudUrl := s.TestBackend.GetCloudUrl()
	mdsUrl := s.TestBackend.GetMdsUrl()
	tests := []saveTest{
		{
			true,
			cloudUrl,
			testBin,
		},
		{
			false,
			mdsUrl,
			testBin,
		},
	}
	for _, tt := range tests {
		var env []string
		if tt.isCloud {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", auth.ConfluentPlatformPassword)}
		}
		configFile := filepath.Join(os.Getenv("HOME"), ".confluent", "config.json")
		// run login to provide context, then logout command and check output
		output := runCommand(s.T(), tt.bin, env, "login -vvvv --save --url "+tt.loginURL, 0, "")
		if tt.isCloud {
			s.Contains(output, loggedInAsWithOrgOutput)
		} else {
			s.Contains(output, loggedInAsOutput)
		}

		got, err := os.ReadFile(configFile)
		s.NoError(err)
		s.Require().Contains(utils.NormalizeNewLines(string(got)), "saved_credentials")

		output = runCommand(s.T(), tt.bin, env, "logout -vvvv", 0, "")
		s.Contains(output, errors.LoggedOutMsg)

		got, err = os.ReadFile(configFile)
		s.NoError(err)
		s.Require().NotContains(utils.NormalizeNewLines(string(got)), "saved_credentials")
	}
}

func (s *CLITestSuite) TestRemoveUsernamePasswordFail() {
	// fail to parse the netrc file should leave it unchanged
	type saveTest struct {
		isCloud  bool
		loginURL string
		bin      string
	}
	cloudUrl := s.TestBackend.GetCloudUrl()
	mdsUrl := s.TestBackend.GetMdsUrl()
	tests := []saveTest{
		{
			true,
			cloudUrl,
			testBin,
		},
		{
			false,
			mdsUrl,
			testBin,
		},
	}
	for _, tt := range tests {
		// store existing credentials in a temp netrc to check that they are not corrupted
		var env []string
		if tt.isCloud {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", auth.ConfluentPlatformPassword)}
		}

		configFile := filepath.Join(os.Getenv("HOME"), ".confluent", "config.json")
		got, err := os.ReadFile(configFile)
		s.NoError(err)
		s.Require().NotContains(utils.NormalizeNewLines(string(got)), "saved_credentials")

		// run login to provide context, then logout command and check output
		runCommand(s.T(), tt.bin, env, "login --url "+tt.loginURL, 0, "") // without save flag so the netrc file won't be modified
		output := runCommand(s.T(), tt.bin, env, "logout", 0, "")
		s.Contains(output, errors.LoggedOutMsg)

		got, err = os.ReadFile(configFile)
		s.NoError(err)
		s.Require().NotContains(utils.NormalizeNewLines(string(got)), "saved_credentials")
	}
}
