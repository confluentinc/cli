package test

import (
	"fmt"
	"strings"

	"github.com/confluentinc/bincover"
)

func (s *CLITestSuite) TestCloudSignup() {
	tests := []CLITest{
		{
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nUS\ny\nConfluent\nPa$$word12\ny\ny\ny\n"))},
			fixture:     "cloud-signup/success.golden",
		},
		{
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nUS\ny\nConfluent\nPa$$word12\nn\ny\nN\nY\nn\ny\n"))},
			fixture:     "cloud-signup/reprompt-on-no-success.golden",
		},
		{
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nBrian\nStrauch\nZZ\nUS\ny\nConfluent\npassword\ny\ny\ny\n"))},
			fixture:     "cloud-signup/bad-country-code.golden",
		},
		{
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("bstrauch@confluent.io\nBrian\nStrauch\nCH\nn\nUS\ny\nConfluent\npassword\ny\ny\ny\n"))},
			fixture:     "cloud-signup/reject-country-code.golden",
		},
		{
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("bstrauch@confluent.io\nBrian\nStrauch\nUS\ny\nConfluent\npassword\nn\ny\ny\ny\n"))},
			fixture:     "cloud-signup/reject-terms-of-service.golden",
		},
		{
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("bstrauch@confluent.io\nBrian\nStrauch\nUS\ny\nConfluent\npassword\ny\nn\ny\ny\n"))},
			fixture:     "cloud-signup/reject-privacy-policy.golden",
		},
		{
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("bstrauch@confluent.io\nBrian\nStrauch\nUS\ny\nConfluent\npassword\ny\ny\nn\ny\n"))},
			fixture:     "cloud-signup/resend-verification-email.golden",
		},
	}

	for _, test := range tests {
		test.args = fmt.Sprintf("cloud-signup --url=%s", s.TestBackend.GetCloudUrl())
		s.runIntegrationTest(test)
	}
}
