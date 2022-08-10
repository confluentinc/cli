package test

func (s *CLITestSuite) TestStreamShare() {
	tests := []CLITest{
		{args: "stream-share provider share list --shared-resource sr-12345", fixture: "stream-share/list-provider-shares.golden"},
		{args: "stream-share provider share describe ss-12345", fixture: "stream-share/describe-provider-share.golden"},
		{args: "stream-share provider share delete ss-12345", fixture: "stream-share/delete-provider-share.golden"},
		{args: "stream-share provider invite create --email user@example.com --environment env-12345 --kafka-cluster lkc-12345 --topic topic-12345", fixture: "stream-share/create-invite.golden"},
		{args: "stream-share provider invite resend ss-12345", fixture: "stream-share/resend-invite.golden"},

		{args: "stream-share consumer share list --shared-resource sr-12345", fixture: "stream-share/list-consumer-shares.golden"},
		{args: "stream-share consumer share delete ss-12345", fixture: "stream-share/delete-consumer-share.golden"},
		{args: "stream-share consumer share describe ss-12345", fixture: "stream-share/describe-consumer-share.golden"},
		{args: "stream-share consumer redeem stream-share-token", fixture: "stream-share/redeem-share.golden"},
		{args: "stream-share consumer redeem stream-share-token --preview", fixture: "stream-share/redeem-preview.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
