package test

func (s *CLITestSuite) TestFeedback() {
	tests := []CLITest{
		{args: "feedback", fixture: "feedback/no-confirm.golden", input: "This CLI is great!\nn\n"},
		{args: "feedback", fixture: "feedback/received.golden", input: "This CLI is great!\ny\n"},
		{args: "feedback", exitCode: 1, fixture: "feedback/too-long.golden", input: "Lorem ipsum dolor sit amet. Qui amet molestiae eum eaque perferendis\ny\n"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
