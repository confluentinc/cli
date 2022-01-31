package test

func (s *CLITestSuite) TestReplicas() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		{args: "kafka replica --help", fixture: "kafka/replicas/help.golden"},
		{args: "kafka replica list -h", fixture: "kafka/replicas/list-help.golden"},
		{args: "kafka replica list --topic topic-exist", fixture: "kafka/replicas/list-topic-replicas.golden"},
		{args: "kafka replica list --topic topic-exist -o json", fixture: "kafka/replicas/list-topic-replicas-json.golden"},
		{args: "kafka replica list --topic topic-exist --partition 2", fixture: "kafka/replicas/list-partition-replicas.golden"},
		{args: "kafka replica list --topic topic-exist --partition 2 -o yaml", fixture: "kafka/replicas/list-partition-replicas-yaml.golden"},
		{args: "kafka replica list", fixture: "kafka/replicas/no-flags-error.golden", wantErrCode: 1},
	}
	for _, tt := range tests {
		tt.login = "default"
		tt.env = []string{"CONFLUENT_REST_URL=" + kafkaRestURL}
		s.runConfluentTest(tt)
	}
}
