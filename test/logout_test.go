package test

import (
	"fmt"
	"os"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (s *CLITestSuite) TestLogout_NoopWhenAlreadyLoggedOut() {
	cloudUrl := s.TestBackend.GetCloudUrl()
	env := []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}

	runCommand(s.T(), testBin, env, "login -vvvv --save --url "+cloudUrl, 0, "")
	runCommand(s.T(), testBin, env, "logout -vvvv", 0, "")

	// Logging out a second time, with no active session, must succeed as a no-op rather than error.
	output := runCommand(s.T(), testBin, env, "logout -vvvv", 0, "")
	s.NotContains(output, "You are now logged out.")
}

func (s *CLITestSuite) TestLogout_NoopAfterReselectingContext() {
	for _, test := range []struct {
		name string
		url  string
		env  []string
	}{
		{
			name: "cloud",
			url:  s.TestBackend.GetCloudUrl(),
			env:  []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)},
		},
		{
			name: "platform",
			url:  s.TestBackend.GetMdsUrl(),
			env:  []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", auth.ConfluentPlatformPassword)},
		},
	} {
		s.Run(test.name, func() {
			resetConfiguration(s.T(), false)
			runCommand(s.T(), testBin, test.env, "login --save --url "+test.url, 0, "")
			cfg := config.New()
			s.Require().NoError(cfg.Load())
			contextName := cfg.CurrentContext
			runCommand(s.T(), testBin, test.env, "logout", 0, "")
			runCommand(s.T(), testBin, test.env, fmt.Sprintf("context use %q", contextName), 0, "")

			output := runCommand(s.T(), testBin, test.env, "logout", 0, "")

			s.Empty(output)
			cfg = config.New()
			s.Require().NoError(cfg.Load())
			s.Empty(cfg.CurrentContext)
			s.NotContains(cfg.SavedCredentials, contextName)
			s.Contains(cfg.Contexts, contextName)
		})
	}
}

func (s *CLITestSuite) TestLogout_RemoveUsernamePassword() {
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
	for _, test := range tests {
		var env []string
		if test.isCloud {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", auth.ConfluentPlatformPassword)}
		}

		// run login to provide context, then logout command and check output
		output := runCommand(s.T(), test.bin, env, "login -vvvv --save --url "+test.loginURL, 0, "")
		if test.isCloud {
			s.Contains(output, loggedInAsWithOrgOutput)
		} else {
			s.Contains(output, loggedInAsOutput)
		}

		got, err := os.ReadFile(config.GetDefaultFilename())
		s.NoError(err)
		s.Require().Contains(utils.NormalizeNewLines(string(got)), "saved_credentials")

		output = runCommand(s.T(), test.bin, env, "logout -vvvv", 0, "")
		s.Contains(output, "You are now logged out.")

		got, err = os.ReadFile(config.GetDefaultFilename())
		s.NoError(err)
		s.Require().NotContains(utils.NormalizeNewLines(string(got)), "saved_credentials")
	}
}

func (s *CLITestSuite) TestLogout_RemoveUsernamePasswordFail() {
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
	for _, test := range tests {
		var env []string
		if test.isCloud {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentCloudEmail), fmt.Sprintf("%s=pass1", auth.ConfluentCloudPassword)}
		} else {
			env = []string{fmt.Sprintf("%s=good@user.com", auth.ConfluentPlatformUsername), fmt.Sprintf("%s=pass1", auth.ConfluentPlatformPassword)}
		}

		got, err := os.ReadFile(config.GetDefaultFilename())
		s.NoError(err)
		s.Require().NotContains(utils.NormalizeNewLines(string(got)), "saved_credentials")

		// run login to provide context, then logout command and check output
		runCommand(s.T(), test.bin, env, "login --url "+test.loginURL, 0, "")
		output := runCommand(s.T(), test.bin, env, "logout", 0, "")
		s.Contains(output, "You are now logged out.")

		got, err = os.ReadFile(config.GetDefaultFilename())
		s.NoError(err)
		s.Require().NotContains(utils.NormalizeNewLines(string(got)), "saved_credentials")
	}
}
