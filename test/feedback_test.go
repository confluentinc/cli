package test

func (s *CLITestSuite) TestFeedback() {
	tests := []CLITest{
		{args: "feedback", fixture: "feedback/no-confirm.golden", input: "n\n"},
		{args: "feedback", fixture: "feedback/received.golden", input: "y\n This CLI is great!\n"},
	}
	for _, tt := range tests {
		s.runIntegrationTest(tt)
	}
}
