package test

func (s *CLITestSuite) TestLocal() {
	tests := []CLITest{
		{args: "local kafka stop", fixture: "local/stop.golden"},
		{args: "local kafka start", fixture: "local/start.golden", regex: true},

		{args: "local kafka stop", fixture: "local/stop.golden"},
	}

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}
