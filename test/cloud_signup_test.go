package test

import (
	"strings"

	"github.com/confluentinc/bincover"
)

func (s *CLITestSuite) TestCloudSignup() {
	tests := []CLITest{
		{
			args:        "cloud-signup --url=" + s.TestBackend.GetCloudUrl(),
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nUS\ny\nConfluent\nPa$$word12\nn\ny\nN\nY\nn\ny\n"))},
			fixture:     "cloud-signup/reprompt-on-no-success.golden",
		},
		{
			args:        "cloud-signup --url=" + s.TestBackend.GetCloudUrl(),
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nUS\ny\nConfluent\nPa$$word12\ny\ny\ny\n"))},
			fixture:     "cloud-signup/success.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}
