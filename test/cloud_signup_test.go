package test

import (
	"fmt"
	"github.com/confluentinc/bincover"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func (s *CLITestSuite) TestCloudSignup_FreeTrialAnnouncement() {
	args := fmt.Sprintf("cloud-signup --url=%s", s.TestBackend.GetCloudUrl())

	s.T().Run("signup only has free trial code", func(tt *testing.T) {
		os.Setenv("HAS_PROMO_CODE_CLAIMS", "onlyFreeTrialCode")

		covCollectorOptions := parseCmdFuncsToCoverageCollectorOptions(
			[]bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nUS\ny\nConfluent\nPa$$word12\ny\ny\ny\n"))},
			[]bincover.PostCmdFunc{})

		output := runCommand(tt, testBin, []string{}, args, 0, covCollectorOptions...)
		require.Contains(tt, output, errors.CloudSignUpMsg)
		require.Contains(tt, output, fmt.Sprintf(errors.FreeTrialSignUpMsg, 400.00))

		os.Unsetenv("HAS_PROMO_CODE_CLAIMS")
	})

	s.T().Run("signup has multiple codes including free trial code", func(tt *testing.T) {
		os.Setenv("HAS_PROMO_CODE_CLAIMS", "multiCodes")

		covCollectorOptions := parseCmdFuncsToCoverageCollectorOptions(
			[]bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nUS\ny\nConfluent\nPa$$word12\ny\ny\ny\n"))},
			[]bincover.PostCmdFunc{})

		output := runCommand(tt, testBin, []string{}, args, 0, covCollectorOptions...)
		require.Contains(tt, output, errors.CloudSignUpMsg)
		// still $400.00, not 420.00
		require.Contains(tt, output, fmt.Sprintf(errors.FreeTrialSignUpMsg, 400.00))

		os.Unsetenv("HAS_PROMO_CODE_CLAIMS")
	})

	s.T().Run("signup missing free trial code", func(tt *testing.T) {
		covCollectorOptions := parseCmdFuncsToCoverageCollectorOptions(
			[]bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("test-signup@confluent.io\nMiles\nTodzo\nUS\ny\nConfluent\nPa$$word12\ny\ny\ny\n"))},
			[]bincover.PostCmdFunc{})

		output := runCommand(tt, testBin, []string{}, args, 0, covCollectorOptions...)
		require.Contains(tt, output, errors.CloudSignUpMsg)
		require.NotContains(tt, output, fmt.Sprintf(errors.FreeTrialSignUpMsg, 400.00))
	})
}

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
