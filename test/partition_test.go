package test

func (s *CLITestSuite) TestPartitions() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		{args: "kafka partition --help", fixture: "kafka/partitions/help.golden"},
		{args: "kafka partition list -h", fixture: "kafka/partitions/list-help.golden"},
		{args: "kafka partition list --topic topic1", fixture: "kafka/partitions/list.golden"},
		{args: "kafka partition list --topic topic1 -o json", fixture: "kafka/partitions/list-json.golden"},
		{args: "kafka partition list --topic topic1 -o yaml", fixture: "kafka/partitions/list-yaml.golden"},
		{args: "kafka partition describe -h", fixture: "kafka/partitions/describe-help.golden"},
		{args: "kafka partition describe 0 --topic topic1", fixture: "kafka/partitions/describe.golden"},
		{args: "kafka partition describe 0 --topic topic1 -o json", fixture: "kafka/partitions/describe-json.golden"},
		{args: "kafka partition describe 0 --topic topic1 -o yaml", fixture: "kafka/partitions/describe-yaml.golden"},
		{args: "kafka partition get-reassignments -h", fixture: "kafka/partitions/reassignments-help.golden"},
		{args: "kafka partition get-reassignments", fixture: "kafka/partitions/reassignments.golden"},
		{args: "kafka partition get-reassignments -o json", fixture: "kafka/partitions/reassignments-json.golden"},
		{args: "kafka partition get-reassignments --topic topic1", fixture: "kafka/partitions/reassignments-by-topic.golden"},
		{args: "kafka partition get-reassignments 0 --topic topic1", fixture: "kafka/partitions/reassignments-by-partition.golden"},
		{args: "kafka partition get-reassignments 0 --topic topic1 -o yaml", fixture: "kafka/partitions/reassignments-by-partition-yaml.golden"},
	}
	for _, tt := range tests {
		tt.login = "default"
		tt.env = []string{"CONFLUENT_REST_URL=" + kafkaRestURL}
		s.runConfluentTest(tt)
	}
}
