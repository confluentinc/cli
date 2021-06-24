package test

import (
	"strings"

	"github.com/confluentinc/bincover"
)

func (s *CLITestSuite) TestSignup() {
	tests := []CLITest{
		{
			Args:        "signup --url=" + s.TestBackend.GetCloudUrl(),
			PreCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nConfluent\nPa$$word12\nn\ny\nN\nY\nn\ny\n"))},
			Fixture:     "signup/signup-reprompt-on-no-success.golden",
		},
		{
			Args:        "signup --url=" + s.TestBackend.GetCloudUrl(),
			PreCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nConfluent\nPa$$word12\ny\ny\ny\n"))},
			Fixture:     "signup/signup-success.golden",
		},
	}

	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}
