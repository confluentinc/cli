package test

func (s *CLITestSuite) TestUpdate() {
	tests := []CLITest{
		{args: "update", fixture: "update/update.golden", input: "y\n"},
		{args: "update", fixture: "update/update-no.golden", input: "n\n"},
		{args: "update --major", fixture: "update/update-major.golden", input: "y\n"},
		{args: "update --no-verify", fixture: "update/update.golden", input: "y\n"},
		{args: "update --yes", fixture: "update/update-yes.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}
