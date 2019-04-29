package test

func (s *CLITestSuite) TestKafkaCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		// Show what commands are available
		{args: "kafka cluster --help", fixture: "kafka-cluster-help.golden"},
		// This is hidden from help, but what if you call it anyway?
		{args: "kafka cluster create", fixture: "kafka1.golden", wantErrCode: 1},
		// This is hidden from help, but what if you call it anyway... with args?
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1", useKafka: "bob", fixture: "kafka2.golden", wantErrCode: 1},
		// This is hidden from help, but what if you call it anyway?
		{args: "kafka cluster delete", fixture: "kafka3.golden", wantErrCode: 1},
		// This is hidden from help, but what if you call it anyway... with args?
		{args: "kafka cluster delete lkc-abc123", fixture: "kafka4.golden", wantErrCode: 1},
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
