package test

func (s *CLITestSuite) TestStreamShareProvider() {
	tests := []CLITest{
		{args: "stream-share provider share list --shared-resource sr-12345", fixture: "stream-share/provider/share/list.golden"},
		{args: "stream-share provider share describe ss-12345", fixture: "stream-share/provider/share/describe.golden"},
		{args: "stream-share provider share delete ss-12345 --force", fixture: "stream-share/provider/share/delete.golden"},
		{args: "stream-share provider share delete ss-12345", input: "y\n", fixture: "stream-share/provider/share/delete-prompt.golden"},
		{args: "stream-share provider share delete ss-12345 ss-12346", fixture: "stream-share/provider/share/delete-multiple-fail.golden", exitCode: 1},
		{args: "stream-share provider share delete ss-12345 ss-54321", input: "n\n", fixture: "stream-share/provider/share/delete-multiple-refuse.golden"},
		{args: "stream-share provider share delete ss-12345 ss-54321", input: "y\n", fixture: "stream-share/provider/share/delete-multiple-success.golden"},

		{args: "stream-share provider invite create --email user@example.com --topic topic-12345 --environment env-123456 --cluster lkc-12345 --schema-registry-subjects sub1,sub2,sub3", fixture: "stream-share/provider/invite/create.golden"},
		{args: "stream-share provider invite resend ss-12345", fixture: "stream-share/provider/invite/resend.golden"},

		{args: "stream-share provider opt-in", fixture: "stream-share/provider/opt-in.golden"},
		{args: "stream-share provider opt-out", input: "y\n", fixture: "stream-share/provider/opt-out-accept.golden"},
		{args: "stream-share provider opt-out", input: "n\n", fixture: "stream-share/provider/opt-out-decline.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestStreamShareConsumer() {
	tests := []CLITest{
		{args: "stream-share consumer share list --shared-resource sr-12345", fixture: "stream-share/consumer/share/list.golden"},
		{args: "stream-share consumer share delete ss-12345 --force", fixture: "stream-share/consumer/share/delete.golden"},
		{args: "stream-share consumer share delete ss-12345", input: "y\n", fixture: "stream-share/consumer/share/delete-prompt.golden"},
		{args: "stream-share consumer share delete ss-12345 ss-12346", fixture: "stream-share/consumer/share/delete-multiple-fail.golden", exitCode: 1},
		{args: "stream-share consumer share delete ss-12345 ss-54321", input: "n\n", fixture: "stream-share/consumer/share/delete-multiple-refuse.golden"},
		{args: "stream-share consumer share delete ss-12345 ss-54321", input: "y\n", fixture: "stream-share/consumer/share/delete-multiple-success.golden"},
		{args: "stream-share consumer share describe ss-12345", fixture: "stream-share/consumer/share/describe.golden"},

		{args: "stream-share consumer redeem stream-share-token", fixture: "stream-share/consumer/redeem.golden"},
		{args: "stream-share consumer redeem stream-share-token --aws-account 111111111111", fixture: "stream-share/consumer/redeem-private-link.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestStreamShare_Autocomplete() {
	tests := []CLITest{
		{args: `__complete stream-share consumer share describe ""`, fixture: "stream-share/consumer/share/describe-autocomplete.golden"},
		{args: `__complete stream-share provider share describe ""`, fixture: "stream-share/provider/share/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
