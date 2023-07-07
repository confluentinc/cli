package test

import "fmt"

func (s *CLITestSuite) TestFeedback() {
	feedback := "Lorem ipsum dolor sit amet. Qui amet molestiae eum eaque perferendis\n"
	tests := []CLITest{
		{args: "feedback", fixture: "feedback/no-confirm.golden", input: "n\n"},
		{args: "feedback", fixture: "feedback/received.golden", input: "y\nThis CLI is great!\n"},
		{args: "feedback", exitCode: 1, fixture: "feedback/too-long.golden", input: fmt.Sprintf("y\n%s", feedback)},
	}
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
