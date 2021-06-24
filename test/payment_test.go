package test

import (
	"strings"

	"github.com/confluentinc/bincover"
)

func (s *CLITestSuite) TestPaymentDescribe() {
	tests := []CLITest{
		{
			Args:    "admin payment describe",
			Fixture: "admin/payment-describe.golden",
		},
	}

	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}

func (s *CLITestSuite) TestPaymentUpdate() {
	tests := []CLITest{
		{
			Args:        "admin payment update",
			PreCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("4242424242424242\n12/70\n999\nBrian Strauch\n"))},
			Fixture:     "admin/payment-update-success.golden",
		},
		{
			Args:        "admin payment update", //testing with CVC failing regex check on first attempt
			PreCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("4242424242424242\n12/70\n99\n999\nBrian Strauch\n"))},
			Fixture:     "admin/payment-update-bad-cvc.golden",
		},
	}
	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}
