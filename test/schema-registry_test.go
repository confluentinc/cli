package test

func (s *CLITestSuite) TestSrCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		// Show what commands are available
		{args: "schema-registry --help", fixture: "schema-registry-help.golden"},
		// This is hidden from help, but what if you call it anyway?
		{args: "schema-registry cluster create", fixture: "sr1.golden", wantErrCode: 1},
		// This is hidden from help, but what if you call it anyway... with args?
		{args: "schema-registry cluster enable --cloud aws --geo us", useKafka: "bob", fixture: "kafka2.golden", wantErrCode: 1},
		// This is hidden from help, but what if you call it anyway?
		{args: "kafka cluster delete", fixture: "kafka3.golden", wantErrCode: 1},
		// This is hidden from help, but what if you call it anyway... with args?
		{args: "kafka cluster delete lkc-abc123", fixture: "kafka4.golden", wantErrCode: 1},
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
