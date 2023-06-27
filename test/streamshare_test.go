package test

func (s *CLITestSuite) TestStreamShare() {
	tests := []CLITest{
		{args: "stream-share provider share list --shared-resource sr-12345", fixture: "stream-share/list-provider-shares.golden"},
		{args: "stream-share provider share describe ss-12345", fixture: "stream-share/describe-provider-share.golden"},
		{args: "stream-share provider share delete ss-12345 --force", fixture: "stream-share/delete-provider-share.golden"},
		{args: "stream-share provider share delete ss-12345", input: "y\n", fixture: "stream-share/delete-provider-share-prompt.golden"},

		{args: "stream-share provider invite create --email user@example.com --topic topic-12345 --environment env-12345 --cluster lkc-12345 --schema-registry-subjects sub1,sub2,sub3", fixture: "stream-share/create-invite.golden"},
		{args: "stream-share provider invite resend ss-12345", fixture: "stream-share/resend-invite.golden"},

		{args: "stream-share provider opt-in", fixture: "stream-share/opt-in.golden"},
		{args: "stream-share provider opt-out", input: "y\n", fixture: "stream-share/opt-out-accept.golden"},
		{args: "stream-share provider opt-out", input: "n\n", fixture: "stream-share/opt-out-decline.golden"},

		{args: "stream-share consumer share list --shared-resource sr-12345", fixture: "stream-share/list-consumer-shares.golden"},
		{args: "stream-share consumer share delete ss-12345 --force", fixture: "stream-share/delete-consumer-share.golden"},
		{args: "stream-share consumer share delete ss-12345", input: "y\n", fixture: "stream-share/delete-consumer-share-prompt.golden"},
		{args: "stream-share consumer share describe ss-12345", fixture: "stream-share/describe-consumer-share.golden"},

		{args: "stream-share consumer redeem stream-share-token", fixture: "stream-share/redeem-share.golden"},
		{args: "stream-share consumer redeem stream-share-token --aws-account-id 111111111111", fixture: "stream-share/redeem-share-private-link.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestStreamShareAutocomplete() {
	tests := []CLITest{
		{args: `__complete stream-share consumer share describe ""`, fixture: "stream-share/describe-consumer-share-autocomplete.golden"},
		{args: `__complete stream-share provider share describe ""`, fixture: "stream-share/describe-provider-share-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
