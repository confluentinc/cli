package test

func (s *CLITestSuite) TestStreamShare() {
	tests := []CLITest{
		{args: "stream-share provider share list", fixture: "stream-share/list-shares.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
