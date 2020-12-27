package test

import (
	"strings"

	"github.com/confluentinc/bincover"
)

func (s *CLITestSuite) TestPaymentDescribe() {
	tests := []CLITest{
		{
			args:    "admin payment describe",
			fixture: "admin/payment-describe.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}

func (s *CLITestSuite) TestPaymentUpdate() {
	tests := []CLITest{
		{
			args:        "admin payment update",
			preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("4242424242424242\n12/70\n999\nBrian Strauch\n"))},
			fixture:     "admin/payment-update-success.golden",
		},
	}
	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}
