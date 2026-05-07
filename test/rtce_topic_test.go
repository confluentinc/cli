package test

func (s *CLITestSuite) TestRtceRtceTopicCreate() {
	tests := []CLITest{
		{args: "rtce rtce-topic create --cloud aws --description \"Customer orders table for real-time analytics\" --region us-west-2 --topic-name orders_topic", fixture: "rtce/rtce-topic/create.golden", useKafka: "lkc-abc123"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestRtceRtceTopicDelete() {
	tests := []CLITest{
		{args: "rtce rtce-topic delete id-1 --force", fixture: "rtce/rtce-topic/delete.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic delete id-1", input: "y\n", fixture: "rtce/rtce-topic/delete-no-force.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic delete id-1 id-2", input: "y\n", fixture: "rtce/rtce-topic/delete-multiple.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic delete invalid", fixture: "rtce/rtce-topic/delete-invalid.golden", exitCode: 1, useKafka: "lkc-abc123"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestRtceRtceTopicDescribe() {
	tests := []CLITest{
		{args: "rtce rtce-topic describe id-1", fixture: "rtce/rtce-topic/describe.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic describe id-1 -o json", fixture: "rtce/rtce-topic/describe-json.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic describe id-1 -o yaml", fixture: "rtce/rtce-topic/describe-yaml.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic describe invalid", fixture: "rtce/rtce-topic/describe-invalid.golden", exitCode: 1, useKafka: "lkc-abc123"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestRtceRtceTopicList() {
	tests := []CLITest{
		{args: "rtce rtce-topic list", fixture: "rtce/rtce-topic/list.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic list --region us-west-2", fixture: "rtce/rtce-topic/list-region.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic list -o json", fixture: "rtce/rtce-topic/list-json.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic list -o yaml", fixture: "rtce/rtce-topic/list-yaml.golden", useKafka: "lkc-abc123"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestRtceRtceTopicUpdate() {
	tests := []CLITest{
		{args: "rtce rtce-topic update id-1 --description \"Customer orders table for real-time analytics\"", fixture: "rtce/rtce-topic/update-description.golden", useKafka: "lkc-abc123"},
		{args: "rtce rtce-topic update invalid", fixture: "rtce/rtce-topic/update-invalid.golden", exitCode: 1, useKafka: "lkc-abc123"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestRtceRtceTopic_Autocomplete() {
	tests := []CLITest{
		{args: "__complete rtce rtce-topic delete \"\"", fixture: "rtce/rtce-topic/delete-autocomplete.golden"},
		{args: "__complete rtce rtce-topic describe \"\"", fixture: "rtce/rtce-topic/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
