package test

func (s *CLITestSuite) TestKafkaCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "kafka cluster --help", fixture: "kafka-cluster-help.golden"},
		{args: "kafka cluster create", useKafka: "bob", fixture: "kafka1.golden", wantErrCode: 1},
		{args: "kafka cluster create anewstart --cloud aws --region us-east-1", useKafka: "bob", fixture: "kafka2.golden", wantErrCode: 1},
		{args: "kafka cluster delete", useKafka: "bob", fixture: "kafka3.golden", wantErrCode: 1},
	}
	resetConfiguration(s.T())
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		s.runTest(tt, serve(s.T()).URL, serveKafkaAPI(s.T()).URL)
	}
}
