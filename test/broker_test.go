package test

func (s *CLITestSuite) TestBroker() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		{args: "kafka broker list -h", fixture: "kafka/broker/list-help.golden"},
		{args: "kafka broker list", fixture: "kafka/broker/list.golden"},
		{args: "kafka broker list -o json", fixture: "kafka/broker/list-json.golden"},
		{args: "kafka broker list -o yaml", fixture: "kafka/broker/list-yaml.golden"},

		{args: "kafka broker describe -h", fixture: "kafka/broker/describe-help.golden"},
		{args: "kafka broker describe 1", fixture: "kafka/broker/describe-1.golden"},
		{args: "kafka broker describe 1 -o json", fixture: "kafka/broker/describe-1-json.golden"},
		{args: "kafka broker describe 1 -o yaml", fixture: "kafka/broker/describe-1-yaml.golden"},
		{args: "kafka broker describe --all", fixture: "kafka/broker/describe-all.golden"},
		{args: "kafka broker describe --all -o json", fixture: "kafka/broker/describe-all-json.golden"},
		{args: "kafka broker describe --all -o yaml", fixture: "kafka/broker/describe-all-yaml.golden"},
		{args: "kafka broker describe 1 --config-name compression.type", fixture: "kafka/broker/describe-1-config.golden"},
		{args: "kafka broker describe --all --config-name compression.type", fixture: "kafka/broker/describe-all-config.golden"},
		{args: "kafka broker describe --all --config-name compression.type -o json", fixture: "kafka/broker/describe-all-config-json.golden"},
		{args: "kafka broker describe 1 --all", wantErrCode: 1, fixture: "kafka/broker/err-all-and-arg.golden"},

		{args: "kafka broker update -h", fixture: "kafka/broker/update-help.golden"},
		{args: "kafka broker update --config compression.type=zip,sasl_mechanism=SASL/PLAIN --all", fixture: "kafka/broker/update-all.golden"},
		{args: "kafka broker update 1 --config compression.type=zip,sasl_mechanism=SASL/PLAIN", fixture: "kafka/broker/update-1.golden"},
		{args: "kafka broker update --config compression.type=zip,sasl_mechanism=SASL/PLAIN", wantErrCode: 1, fixture: "kafka/broker/err-need-all-or-arg.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		tt.env = []string{"CONFLUENT_REST_URL=" + kafkaRestURL}
		s.runConfluentTest(tt)
	}
}
