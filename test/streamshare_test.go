package test

func (s *CLITestSuite) TestStreamShare() {
	tests := []CLITest{
		{args: "stream-share provider share list --shared-resource sr-12345", fixture: "stream-share/list-provider-shares.golden"},
		{args: "stream-share provider share describe ss-12345", fixture: "stream-share/describe-provider-share.golden"},
		{args: "stream-share provider share delete ss-12345", fixture: "stream-share/delete-provider-share.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
