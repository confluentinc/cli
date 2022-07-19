package test

func (s *CLITestSuite) TestStreamShare() {
	tests := []CLITest{
		{args: "stream-share provider share list", fixture: "stream-share/list-provider-shares.golden"},
		{args: "stream-share provider share describe ss-12345", fixture: "stream-share/describe-provider-share.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
