package test

import (
	"strings"

	"github.com/confluentinc/bincover"
)

func (s *CLITestSuite) TestSignup() {
	tests := []CLITest{
		{
			args:        "signup --url=" + s.TestBackend.GetCloudUrl(),
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("already-exists@confluent.io\nMiles\nTodzo\nConfluent\nPa$$word12\ny\ny\ny\n"))},
			fixture:     "signup/signup-email-exists.golden",
		},
		{
			args:        "signup --url=" + s.TestBackend.GetCloudUrl(),
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nConfluent\nPa$$word12\nn\ny\nN\nY\nn\ny\n"))},
			fixture:     "signup/signup-reprompt-on-no-success.golden",
		},
		{
			args:        "signup --url=" + s.TestBackend.GetCloudUrl(),
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nConfluent\nPa$$word12\ny\ny\ny\n"))},
			fixture:     "signup/signup-success.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}
