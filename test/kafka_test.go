package test

func (s *CLITestSuite) TestKafkaCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		//{args: "kafka cluster --help", fixture: "kafka-cluster-help.golden"},
		//{args: "kafka cluster create", fixture: "kafka1.golden", wantErrCode: 1},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 -vvv", useKafka: "bob", fixture: "kafka2.golden", wantErrCode: 1},
		//{args: "kafka cluster delete", fixture: "kafka3.golden", wantErrCode: 1},
		//{args: "kafka cluster delete lkc-abc123", fixture: "kafka4.golden", wantErrCode: 1},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL, kafkaAPIURL)
	}
}
