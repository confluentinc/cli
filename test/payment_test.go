package test

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
