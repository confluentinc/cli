package test

func (s *CLITestSuite) TestConnectCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		// Show what commands are available
		{args: "connector --help", fixture: "connector-help.golden"},
		{args: "connector-catalog --help", fixture: "connector-catalog-help.golden"},
		{args: "environment use a-595", fixture: "kafka0.golden", wantErrCode: 0},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1", fixture: "kafka2.golden", wantErrCode: 0},
		{args: "connector create --config input/connector-config.yaml", fixture: "connector-update.golden"},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		urlVal := serve(s.T(), kafkaAPIURL)
		s.runCcloudTest(tt, urlVal.URL, kafkaAPIURL)
	}
}
