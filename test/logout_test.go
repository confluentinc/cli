package test

import (
	"fmt"
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
	removedFromNetrcOutput = `Removed credentials for user "good@user.com" from netrc file "netrc_test"`
	loggedOutOutput        = fmt.Sprintf(errors.LoggedOutMsg)
)

func (s *CLITestSuite) TestRemoveUsernamePassword() {
	type saveTest struct {
		input    string
		isCloud  bool
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
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "logout", "netrc-remove"),
			true,
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", "logout", "netrc-remove-username-password.golden"),
			cloudUrl,
			testBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "logout", "netrc-remove"),
			false,
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", "logout", "netrc-remove-username-password.golden"),
			mdsUrl,
			testBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "logout", "netrc-empty"),
			true,
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", "logout", "empty.golden"),
			cloudUrl,
			testBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "logout", "netrc-empty"),
			false,
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "output", "logout", "empty.golden"),
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
		originalNetrc, err := os.ReadFile(tt.input)
		s.NoError(err)
		err = os.WriteFile(netrc.NetrcIntegrationTestFile, originalNetrc, 0600)
		s.NoError(err)

		// run login to provide context, then logout command and check output
		output := runCommand(s.T(), tt.bin, env, "login -vvvv --save --url "+tt.loginURL, 0)
		s.Contains(output, savedToNetrcOutput)
		if tt.isCloud {
			s.Contains(output, loggedInAsWithOrgOutput)
		} else {
			s.Contains(output, loggedInAsOutput)
		}

		output = runCommand(s.T(), tt.bin, env, "logout -vvvv", 0)
		s.Contains(output, loggedOutOutput)
		s.Contains(output, removedFromNetrcOutput)

		// check netrc file matches wanted file
		got, err := os.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantBytes, err := os.ReadFile(tt.wantFile)
		s.NoError(err)
		s.Equal(utils.NormalizeNewLines(string(wantBytes)), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}

func (s *CLITestSuite) TestRemoveUsernamePasswordFail() {
	// fail to parse the netrc file should leave it unchanged
	type saveTest struct {
		input    string
		isCloud  bool
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
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "logout", "netrc-remove-ccloud-fail"),
			true,
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "logout", "netrc-remove-ccloud-fail"),
			cloudUrl,
			testBin,
		},
		{
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "logout", "netrc-remove-mds-fail"),
			false,
			filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "logout", "netrc-remove-mds-fail"),
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
		originalNetrc, err := os.ReadFile(tt.input)
		s.NoError(err)
		original := strings.Replace(string(originalNetrc), urlPlaceHolder, tt.loginURL, 1)
		err = os.WriteFile(netrc.NetrcIntegrationTestFile, []byte(original), 0600)
		s.NoError(err)

		// run login to provide context, then logout command and check output
		runCommand(s.T(), tt.bin, env, "login --url "+tt.loginURL, 0) // without save flag so the netrc file won't be modified
		output := runCommand(s.T(), tt.bin, env, "logout", 0)
		s.Contains(output, loggedOutOutput)

		// check netrc file matches wanted file
		got, err := os.ReadFile(netrc.NetrcIntegrationTestFile)
		s.NoError(err)
		wantBytes, err := os.ReadFile(tt.wantFile)
		s.NoError(err)
		wantString := strings.Replace(string(wantBytes), urlPlaceHolder, tt.loginURL, 1)
		s.Equal(utils.NormalizeNewLines(wantString), utils.NormalizeNewLines(string(got)))
	}
	_ = os.Remove(netrc.NetrcIntegrationTestFile)
}
