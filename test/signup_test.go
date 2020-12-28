package test

func (s *CLITestSuite) TestSignup() {
	tests := []CLITest{
		{
			args:    "signup",
			fixture: "signup-success.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}
